package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"github.com/registryx/registryx/backend/pkg/api"
	"github.com/registryx/registryx/backend/pkg/audit"
	"github.com/registryx/registryx/backend/pkg/auth"
	"github.com/registryx/registryx/backend/pkg/config"
	"github.com/registryx/registryx/backend/pkg/costs"
	"github.com/registryx/registryx/backend/pkg/database"
	"github.com/registryx/registryx/backend/pkg/email"
	"github.com/registryx/registryx/backend/pkg/intelligence"
	"github.com/registryx/registryx/backend/pkg/metadata"
	"github.com/registryx/registryx/backend/pkg/middleware"
	"github.com/registryx/registryx/backend/pkg/policy"
	"github.com/registryx/registryx/backend/pkg/queue"
	"github.com/registryx/registryx/backend/pkg/registry"
	"github.com/registryx/registryx/backend/pkg/scanner"
	"github.com/registryx/registryx/backend/pkg/storage"
	"github.com/registryx/registryx/backend/pkg/webhook"
	"github.com/registryx/registryx/backend/pkg/license"
)

func main() {
	cfg := config.Load()
	
	// Initialize License System (Early ENV check)
	if err := license.Init(); err != nil {
		log.Printf("License Init Warning: %v", err)
	}


	fmt.Printf("Starting RegistryX Backend (VERSION 2.5) on %s...\n", cfg.ServerPort)

	// Initialize Storage
	store, err := storage.NewS3Driver(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage driver: %v", err)
	}

	// Initialize Database with Retry
	var dbConn *sql.DB
	for i := 0; i < 10; i++ {
		dbConn, err = database.Connect(cfg)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database (attempt %d/10): %v. Retrying in 2s...", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to database after retries: %v", err)
	}

	// Dynamic License Check (DB Fallback)
	license.SetDB(dbConn)
	if err := license.LoadFromDB(); err != nil {
		log.Printf("License DB Load Warning: %v", err)
	}


	// Initialize Metadata Service
	metaService := metadata.NewService(dbConn)

	// Initialize Webhook Service (Notifications)
	webhookService := webhook.NewService(dbConn)


	// Initialize Scanner Service (with webhook support)
	scanService := scanner.NewService(dbConn, cfg, webhookService)

	// Initialize Policy Service
	policyService := policy.NewService(dbConn)

	queueService, err := queue.NewService(cfg)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis Queue: %v. Async scanning will be disabled.\n", err)
	}

	// 12. Intelligence Service (EPSS Vulnerability Prioritization)
	intelService := intelligence.NewService(dbConn)

	// 7. Start Background Worker
	if queueService != nil {
		go func() {
			log.Println("Starting Scan Worker...")
			for {
				job, err := queueService.DequeueScan(context.Background())
				if err != nil {
					log.Printf("Worker Queue Error: %v\n", err)
					time.Sleep(5 * time.Second) // Backoff
					continue
				}
				
				log.Printf("Worker: Processing scan for %s (Repo: %s)\n", job.Reference, job.Repository)
				scanService.ScanManifest(context.Background(), job.ManifestID, job.Repository, job.Reference)
				
				// 3. Enrich with Intelligence Priorities
				_ = intelService.CalculateManifestPriorities(context.Background(), job.ManifestID)

				// 4. Recalculate health score after scan
				metaService.CalculateAndStoreHealthScore(context.Background(), job.ManifestID)
				
				log.Printf("Worker: Scan finished for %s\n", job.Reference)
			}
		}()

		// Start Periodic EPSS Intelligence Refresh (Daily)
		go func() {
			log.Println("Starting Intelligence Refresh Worker (Bulk EPSS)...")
			for {
				// Wait 24 hours between refreshes
				// For first run, wait a bit to let system settle
				time.Sleep(1 * time.Hour) 
				
				log.Println("[Intelligence] Starting periodic EPSS data refresh...")
				err := intelService.RefreshEPSSData(context.Background())
				if err != nil {
					log.Printf("[Intelligence] Refresh failed: %v\n", err)
				}
				
				time.Sleep(23 * time.Hour)
			}
		}()
	}

	// 9. Email Service
	emailService := email.NewService(cfg)
	
	// 10. Audit Service
	auditService := audit.NewService(dbConn)

	// 11. Auth Service (Service Accounts + Sessions)
	var redisClient *redis.Client
	if queueService != nil {
		redisClient = queueService.Client
	}
	authService := auth.NewService(dbConn, emailService, auditService, redisClient, cfg.JWTSecret, cfg)


	costConfig := &costs.CostConfig{
		StorageCostPerGBMonth: cfg.StorageCostPerGBMonth, 
		BandwidthCostPerGB:    cfg.BandwidthCostPerGB, 
		RegistryRegion:        "custom",
		CostMode:              cfg.CostMode,
		StorageCapacityTB:     cfg.StorageCapacityTB,
	}
	costService := costs.NewService(dbConn, costConfig)

	// Initialize Registry Handler
	regHandler := registry.NewHandler(cfg, store, metaService, scanService, policyService, queueService, webhookService, auditService)
	
	// Initialize Dashboard Handler
	dashHandler := api.NewDashboardHandler(metaService, scanService, policyService, authService, store, cfg, auditService, webhookService)

	// Initialize Advanced Features Handler
	advancedHandler := api.NewAdvancedHandler(intelService, costService)


	// Router Setup (Gorilla Mux)
	r := mux.NewRouter()

	// Middleware
	authMiddleware := middleware.AuthMiddleware(cfg.JWTSecret, redisClient, cfg.BackendURL)

	// Dashboard API Group
	apiV1 := r.PathPrefix("/api/v1").Subrouter()
	apiV1.Handle("/stats", authMiddleware(http.HandlerFunc(dashHandler.GetStats))).Methods("GET")
	apiV1.HandleFunc("/service-accounts", dashHandler.ListServiceAccounts).Methods("GET")
	apiV1.HandleFunc("/service-accounts", dashHandler.CreateServiceAccount).Methods("POST")
	apiV1.HandleFunc("/service-accounts/{id}", dashHandler.RevokeServiceAccount).Methods("DELETE")
	apiV1.Handle("/dependencies", authMiddleware(http.HandlerFunc(dashHandler.GetDependencyGraph))).Methods("GET")

	// Auth API
	apiV1.HandleFunc("/auth/register", dashHandler.Register).Methods("POST")
	apiV1.HandleFunc("/auth/token", authService.TokenHandler).Methods("GET")
	apiV1.HandleFunc("/auth/login", dashHandler.Login).Methods("POST")
	apiV1.Handle("/auth/logout", authMiddleware(http.HandlerFunc(dashHandler.Logout))).Methods("POST")
    apiV1.HandleFunc("/auth/forgot-password", dashHandler.ForgotPassword).Methods("POST")
	apiV1.HandleFunc("/auth/reset-with-key", dashHandler.ResetPasswordWithKey).Methods("POST")
	apiV1.HandleFunc("/auth/reset-password", dashHandler.ResetPassword).Methods("POST")
	
	apiV1.Handle("/auth/change-password", authMiddleware(http.HandlerFunc(dashHandler.ChangePassword))).Methods("POST")
	apiV1.Handle("/user/audit-logs", authMiddleware(http.HandlerFunc(dashHandler.GetAuditLogs))).Methods("GET")
	
	// Admin / System
	apiV1.Handle("/system/sessions", authMiddleware(http.HandlerFunc(dashHandler.GetActiveSessions))).Methods("GET")
	apiV1.Handle("/system/sessions/{id}", authMiddleware(http.HandlerFunc(dashHandler.RevokeSession))).Methods("DELETE")
	
	// System API
	apiV1.HandleFunc("/health-check", dashHandler.HealthCheck).Methods("GET") // Added health-check
	apiV1.HandleFunc("/policy", dashHandler.GetPolicy).Methods("GET")
	apiV1.HandleFunc("/policy", dashHandler.UpdatePolicy).Methods("PUT")
	apiV1.Handle("/system/security/policy", authMiddleware(http.HandlerFunc(dashHandler.GetSecurityPolicy))).Methods("GET")
	apiV1.Handle("/system/security/policy", authMiddleware(http.HandlerFunc(dashHandler.UpdateSecurityPolicy))).Methods("PUT")
	
	apiV1.Handle("/system/security/policy/overrides", authMiddleware(http.HandlerFunc(dashHandler.ListRepositorySecurityPolicies))).Methods("GET")
	apiV1.Handle("/system/security/policy/overrides", authMiddleware(http.HandlerFunc(dashHandler.UpdateRepositorySecurityPolicy))).Methods("POST")
	apiV1.Handle("/system/security/policy/overrides/{repository:.+}", authMiddleware(http.HandlerFunc(dashHandler.DeleteRepositorySecurityPolicy))).Methods("DELETE")
	
	apiV1.Handle("/repositories", authMiddleware(http.HandlerFunc(dashHandler.CreateRepository))).Methods("POST")
	
	// System / Admin
	apiV1.HandleFunc("/system/config", dashHandler.GetSystemConfig).Methods("GET") // Expose config
	apiV1.Handle("/system/license", authMiddleware(http.HandlerFunc(dashHandler.UpdateLicense))).Methods("POST")
	apiV1.Handle("/system/gc", authMiddleware(http.HandlerFunc(dashHandler.GarbageCollect))).Methods("POST")

	// Webhooks
	apiV1.Handle("/system/webhooks", authMiddleware(http.HandlerFunc(dashHandler.ListWebhooks))).Methods("GET")
	apiV1.Handle("/system/webhooks", authMiddleware(http.HandlerFunc(dashHandler.CreateWebhook))).Methods("POST")
	apiV1.Handle("/system/webhooks/{id}", authMiddleware(http.HandlerFunc(dashHandler.DeleteWebhook))).Methods("DELETE")
	apiV1.Handle("/system/webhooks/{id}/test", authMiddleware(http.HandlerFunc(dashHandler.TestWebhook))).Methods("POST")

	// User Management (Admin)
	apiV1.Handle("/users", authMiddleware(http.HandlerFunc(dashHandler.ListUsers))).Methods("GET")
	apiV1.Handle("/users", authMiddleware(http.HandlerFunc(dashHandler.InviteUser))).Methods("POST")
	apiV1.Handle("/users/{id}", authMiddleware(http.HandlerFunc(dashHandler.DeleteUser))).Methods("DELETE")
	apiV1.Handle("/users/{id}/role", authMiddleware(http.HandlerFunc(dashHandler.UpdateUserRole))).Methods("PUT")
	
	// Specific routes must come BEFORE greedy routes matches
	// Specific routes must come BEFORE greedy routes matches
	// We need to match {name} up to "/tags/" or "/manifests/"
	// Since {name} can contain slashes, we need careful ordering.
	// But actually, just put specific ones first and Mux should handle it if patterns differ.
	// The problem is {name:.+} matches everything.
	// Let's force it to not match if it contains /tags/ or /manifests/ ? No, regex is hard here.
	
	// Better approach: Use a router sub-path or specific matching order.
	// Gorilla Mux matches in order.
	
	apiV1.HandleFunc("/repositories/{name:.+}/tags/{tag}", dashHandler.DeleteTag).Methods("DELETE")
	
	// FIX: Use a regex that explicitly stops at /manifests/
	// This is tricky because {name} is greedy.
	// Let's try matching manifests route explicitly with strict path.
	apiV1.HandleFunc("/repositories/{name:.+}/manifests/{reference}", dashHandler.DeleteManifest).Methods("DELETE")
	apiV1.HandleFunc("/repositories/{name:.+}/manifests/{reference}", dashHandler.GetManifestDetails).Methods("GET")
	
	// Scan-related routes
	apiV1.HandleFunc("/repositories/{name:.+}/manifests/{reference}/scan/status", dashHandler.GetScanStatus).Methods("GET")
	apiV1.HandleFunc("/repositories/{name:.+}/manifests/{reference}/scan/report", dashHandler.DownloadScanReport).Methods("GET")
	apiV1.HandleFunc("/repositories/{name:.+}/manifests/{reference}/scan/history", dashHandler.GetScanHistory).Methods("GET")
	apiV1.HandleFunc("/repositories/{name:.+}/manifests/{reference}/scan/trigger", dashHandler.TriggerManualScan).Methods("POST")
	
	// Greedy match for repository name - MUST BE LAST
	// Use MatcherFunc to ensure we don't accidentally match /manifests/ or /tags/
	// because {name:.+} is very greedy.
	apiV1.Handle("/repositories/{name:.+}", authMiddleware(http.HandlerFunc(dashHandler.DeleteRepository))).Methods("DELETE").MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		return !strings.Contains(r.URL.Path, "/manifests/") && !strings.Contains(r.URL.Path, "/tags/")
	})

	// Advanced Features API
	apiV1.HandleFunc("/vulnerabilities/prioritized", advancedHandler.GetPrioritizedVulnerabilities).Methods("GET")
	apiV1.HandleFunc("/vulnerabilities/intelligence/{cve}", advancedHandler.GetVulnIntelligence).Methods("GET")
	apiV1.HandleFunc("/vulnerabilities/refresh-epss", advancedHandler.RefreshEPSS).Methods("POST")
	


	apiV1.Handle("/costs/dashboard", authMiddleware(http.HandlerFunc(advancedHandler.GetCostDashboard))).Methods("GET")
	apiV1.Handle("/costs/zombie-images", authMiddleware(http.HandlerFunc(advancedHandler.GetZombieImages))).Methods("GET")
	apiV1.Handle("/costs/refresh", authMiddleware(http.HandlerFunc(advancedHandler.RefreshCosts))).Methods("POST")
	apiV1.Handle("/costs/cleanup-zombies", authMiddleware(http.HandlerFunc(advancedHandler.CleanupZombies))).Methods("POST")

	// Auth Service
	r.HandleFunc("/auth/token", authService.TokenHandler).Methods("GET")

	// Middleware (Already declared above)
	// authMiddleware := middleware.AuthMiddleware

	// OCI V2 Distribution API
	v2 := r.PathPrefix("/v2").Subrouter()
	// Apply Middleware? For granular control we wrap handlers.
	
	// Base
	v2.Handle("/", http.HandlerFunc(regHandler.BaseCheck)).Methods("GET")
	
	// Blobs
	// Check Blob (HEAD)
	// {name:.+} matches "repo/subrepo"
	v2.Handle("/{name:.+}/blobs/{digest}", authMiddleware(http.HandlerFunc(regHandler.CheckBlob))).Methods("HEAD")
	v2.Handle("/{name:.+}/blobs/{digest}", authMiddleware(http.HandlerFunc(regHandler.GetBlob))).Methods("GET")

	// Start Upload (POST)
	v2.Handle("/{name:.+}/blobs/uploads/", authMiddleware(http.HandlerFunc(regHandler.StartBlobUpload))).Methods("POST")
	
	// Patch Upload (PATCH)
	v2.Handle("/{name:.+}/blobs/uploads/{uuid}", authMiddleware(http.HandlerFunc(regHandler.PatchBlobData))).Methods("PATCH")
	
	// Finish Upload (PUT)
	v2.Handle("/{name:.+}/blobs/uploads/{uuid}", authMiddleware(http.HandlerFunc(regHandler.PutBlobUpload))).Methods("PUT")

	// Manifests Management
	v2.Handle("/{name:.+}/manifests/{reference}", http.HandlerFunc(regHandler.GetManifest)).Methods("GET", "HEAD")
	v2.Handle("/{name:.+}/manifests/{reference}", authMiddleware(http.HandlerFunc(regHandler.PutManifest))).Methods("PUT")
	
	// Tags List
	v2.Handle("/{name:.+}/tags/list", authMiddleware(http.HandlerFunc(regHandler.Tags))).Methods("GET")
	
	// Catalog (Listing Repos) - Public for GUI MVP
	v2.Handle("/_catalog", authMiddleware(http.HandlerFunc(regHandler.Catalog))).Methods("GET")

	// Global Middleware Function
	globalMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Log Request
			log.Printf("Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

			// CORS Headers (Production Tighter)
			origin := r.Header.Get("Origin")
			if origin != "" {
				// Check against allowed origins
				allowed := false
				for _, o := range strings.Split(cfg.CORSAllowedOrigins, ",") {
					if strings.TrimSpace(o) == "*" || strings.TrimSpace(o) == origin {
						allowed = true
						break
					}
				}

				if allowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Docker-Upload-UUID, X-Requested-With")
			
			// Handle Preflight
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}

	// Start Server with Graceful Shutdown
	srv := &http.Server{
		Addr:    cfg.ServerPort,
		Handler: globalMiddleware(r),
	}

	// Run server in a goroutine so that it doesn't block
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Println("Server Started")

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 30 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 30 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
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
)

func main() {
	cfg := config.Load()
	fmt.Printf("Starting RegistryX Backend (VERSION 2.2 - HEALTH ALGO UPDATE) on %s...\n", cfg.ServerPort)

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

	// Initialize Metadata Service
	metaService := metadata.NewService(dbConn)

	// Initialize Scanner Service
	scanService := scanner.NewService(dbConn, cfg)

	// Initialize Policy Service
	policyService := policy.NewService()

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

	// 8. Webhook Service
	webhookService := webhook.NewService(cfg.WebhookURL)

	// 9. Email Service
	emailService := email.NewService(cfg)
	
	// 10. Audit Service
	auditService := audit.NewService(dbConn)

	// 11. Auth Service (Service Accounts + Sessions)
	var redisClient *redis.Client
	if queueService != nil {
		redisClient = queueService.Client
	}
	authService := auth.NewService(dbConn, emailService, auditService, redisClient, cfg.JWTSecret)


	costConfig := &costs.CostConfig{
		StorageCostPerGBMonth: cfg.StorageCostPerGBMonth, 
		BandwidthCostPerGB:    cfg.BandwidthCostPerGB, 
		RegistryRegion:        "custom",
	}
	costService := costs.NewService(dbConn, costConfig)

	// Initialize Registry Handler
	regHandler := registry.NewHandler(cfg, store, metaService, scanService, policyService, queueService, webhookService, auditService)
	
	// Initialize Dashboard Handler
	dashHandler := api.NewDashboardHandler(metaService, scanService, policyService, authService, store, cfg, auditService)

	// Initialize Advanced Features Handler
	advancedHandler := api.NewAdvancedHandler(intelService, costService)

	// Router Setup (Gorilla Mux)
	r := mux.NewRouter()

	// Middleware
	authMiddleware := middleware.AuthMiddleware(cfg.JWTSecret, redisClient)

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
	
	apiV1.Handle("/repositories", authMiddleware(http.HandlerFunc(dashHandler.CreateRepository))).Methods("POST")
	
	// System / Admin
	apiV1.HandleFunc("/system/config", dashHandler.GetSystemConfig).Methods("GET") // Expose config
	apiV1.Handle("/system/gc", authMiddleware(http.HandlerFunc(dashHandler.GarbageCollect))).Methods("POST")
	
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
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
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

	// Start Server with Global Middleware
	log.Fatal(http.ListenAndServe(cfg.ServerPort, globalMiddleware(r)))
}

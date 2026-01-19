package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	UserKey     ContextKey = "user"
	UsernameKey ContextKey = "username"
	RoleKey     ContextKey = "role"
	AccessKey   ContextKey = "access"
	SessionIDKey ContextKey = "session_id"
)

// AuthMiddleware handles Docker Registry authentication challenges.
func AuthMiddleware(jwtSecret string, rdb *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Debug Log
		log.Printf("[AuthMiddleware] Intercepting: %s\n", r.URL.Path)

		// 1. Skip auth for /v2/ base check if we want to allow anonymous discovery,
		// but typically we want to challenge everything except the auth endpoint itself.
		// The /auth/token endpoint is NOT wrapped by this middleware in main.go.

		// Bypass for internal scanner (localhost)
		// RemoteAddr examples: "127.0.0.1:12345", "[::1]:12345"
		if strings.HasPrefix(r.RemoteAddr, "127.0.0.1:") || strings.HasPrefix(r.RemoteAddr, "[::1]:") {
			fmt.Printf("[AuthMiddleware] Allowing internal request from %s\n", r.RemoteAddr)
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			sendChallenge(w, r)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		
		// 2. Parse and Validate Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate algo
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// Use the provided secret
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			fmt.Printf("Invalid token: %v\n", err)
			sendChallenge(w, r)
			return
		}

		// 3. Extract Claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// --- Session Verification ---
			if rdb != nil {
				// We expect a 'jti' (JWT ID) in the claims for session tracking
				sid, _ := claims["jti"].(string)
				
				// For Docker tokens that don't have JTI (e.g. from /auth/token request), 
				// we might allow them if they are short-lived.
				// But for Dashboard/UI login, we check Redis.
				if sid != "" {
					exists, err := rdb.Exists(r.Context(), "session:"+sid).Result()
					if err != nil || exists == 0 {
						fmt.Printf("[Auth] Session %s expired or revoked\n", sid)
						sendChallenge(w, r)
						return
					}
					// Update last active
					rdb.Expire(r.Context(), "session:"+sid, 24*time.Hour)
				}
			}

			// Inject into context
			ctx := context.WithValue(r.Context(), UserKey, claims["sub"])
			ctx = context.WithValue(ctx, UsernameKey, claims["username"])
			ctx = context.WithValue(ctx, RoleKey, claims["role"])
			
			if sid, ok := claims["jti"].(string); ok {
				ctx = context.WithValue(ctx, SessionIDKey, sid)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			sendChallenge(w, r)
		}
	})
	}
}

// sendChallenge returns the 401 header that tells Docker where to get a token.
func sendChallenge(w http.ResponseWriter, r *http.Request) {
	// Construct the realm URL (assuming localhost:5000 for now)
	// scope should match the request (e.g. repository:my-image:pull)
	// We need to construct the scope string based on the request URL.
	// URL Pattern: /v2/<name>/...
	
	scope := ""
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	
	// Basic scope deduction logic
	if len(pathParts) > 2 && pathParts[0] == "v2" {
		// part[1] could be the repo name
		// But repo names can be namespaced (foo/bar).
		// We'll check the action.
		
		// Simplify: just say "repository:catalog:*" or similar if we can't parse it easily yet.
		// For proper challenge, we try to guess.
		// If path is /v2/alpine/blobs/..., repo is alpine.
		
		// For MVP, empty scope triggers a generic login, which is often enough for the client to retry with *some* scope.
		// Docker client usually knows what it wants and sends the scope in the /auth/token request parameter *after* receiving this 401.
		// The 'scope' in the Www-Authenticate header is what we *require*.
	}

	// Dynamic realm
	realm := "http://localhost:5000/auth/token"
	service := "registryx"
	
	authHeader := fmt.Sprintf(`Bearer realm="%s",service="%s"`, realm, service)
	if scope != "" {
		authHeader += fmt.Sprintf(`,scope="%s"`, scope)
	}

	w.Header().Set("Www-Authenticate", authHeader)
	w.Header().Set("Docker-Distribution-Api-Version", "registry/2.0")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"errors": [{"code": "UNAUTHORIZED", "message": "authentication required"}]}`))
}

package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenResponse is the JSON response for a successful token request.
type TokenResponse struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token"` // Docker client likes both
	ExpiresIn   int    `json:"expires_in"`
	IssuedAt    string `json:"issued_at"`
}

// Access describes the resource action being requested.
type Access struct {
	Type    string   `json:"type"`    // e.g. "repository"
	Name    string   `json:"name"`    // e.g. "alpine"
	Actions []string `json:"actions"` // e.g. ["pull", "push"]
}

// TokenHandler implements GET /auth/token
func (s *Service) TokenHandler(w http.ResponseWriter, r *http.Request) {
	service := r.URL.Query().Get("service")
	scope := r.URL.Query().Get("scope")
	
	// 1. Authenticate the user (Basic Auth)
	rawUser, rawPass, hasAuth := r.BasicAuth()
	username := "anonymous"
	subject := "anonymous"
	
	if hasAuth {
		validUser, err := s.ValidateCredentials(r.Context(), rawUser, rawPass)
		if err != nil {
			fmt.Printf("Auth failed for user %s: %v\n", rawUser, err)
			w.Header().Set("Www-Authenticate", `Bearer realm="http://localhost:5000/auth/token",service="registryx"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		username = validUser.Username
		subject = validUser.ID.String()
		fmt.Printf("Auth request verified for user: %s (ID: %s)\n", username, subject)
	}

	// 2. Parse Requested Access
	access := parseScope(scope)

	// 3. Authorize Access (RBAC)
	grantedAccess := []*Access{}
	
	for _, a := range access {
		if a.Type == "repository" {
			newActions := []string{}
			
			// Parse Namespace
			parts := strings.SplitN(a.Name, "/", 2)
			namespace := "library"
			if len(parts) == 2 {
				namespace = parts[0]
			}
			
			// Determine Permissions
			canPull := false
			canPush := false
			
			if username == "admin" {
				canPull = true
				canPush = true
			} else if username == namespace {
				canPull = true
				canPush = true
			} else if namespace == "library" {
				canPull = true
				canPush = true // Every user can push to library privately
			}

			for _, action := range a.Actions {
				if action == "pull" && canPull {
					newActions = append(newActions, "pull")
				} else if action == "push" && canPush {
					newActions = append(newActions, "push")
				}
			}
			
			if len(newActions) > 0 {
				grantedAccess = append(grantedAccess, &Access{
					Type:    a.Type,
					Name:    a.Name,
					Actions: newActions,
				})
			}
		}
	}

	// 4. Generate JWT
	tokenString, err := s.generateRegistryToken(service, subject, grantedAccess)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	resp := TokenResponse{
		Token:       tokenString,
		AccessToken: tokenString,
		ExpiresIn:   3600,
		IssuedAt:    time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// parseScope parses "repository:samalba/my-app:pull,push"
func parseScope(scope string) []*Access {
	if scope == "" {
		return []*Access{}
	}
	parts := strings.Split(scope, ":")
	if len(parts) < 3 {
		return []*Access{}
	}
	
	// Handle names that might have colons? Docker spec says type:name:action
	// But name can contain slashes.
	// Standard format: type:name:action1,action2
	resType := parts[0]
	resName := strings.Join(parts[1:len(parts)-1], ":") // Join middle parts just in case
	resActions := strings.Split(parts[len(parts)-1], ",")
	
	return []*Access{&Access{
		Type:    resType,
		Name:    resName,
		Actions: resActions,
	}}
}

// generateToken signs a JWT
// Note: In real prod, use a persistent RSA Private Key. 
// For this MVP session, we'll generate a random key on startup or use a static secret (HMAC) for simplicity
// BUT Docker requires RS256 usually if checking signatures against a public key derived from it.
// We will use HS256 for internal verification if we are the only ones checking it.
// However, if we want to be correct, we need a signing key. Let's use a dummy secret for now.
func (s *Service) generateRegistryToken(service, subject string, access []*Access) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":    "registryx-auth",
		"sub":    subject,
		"aud":    service,
		"exp":    now.Add(time.Hour).Unix(),
		"nbf":    now.Unix(),
		"iat":    now.Unix(),
		"access": access,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString([]byte(s.JWTSecret))
}

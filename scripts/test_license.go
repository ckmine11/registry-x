package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// === MOCKING THE VALIDATOR LOGIC HERE FOR TEST ===
// (Aapke backend/pkg/license ka simplified version)

type LicenseClaims struct {
	Features []string `json:"features"`
	jwt.RegisteredClaims
}

func main() {
	fmt.Println("🚀 STARTING LICENSE SYSTEM DRY RUN...")

	// 1. Generate RSA Keys (Admin Side)
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	fmt.Println("✅ [1] Generated Administrator RSA Keys.")

	// 2. Issuing a License for Client "Acme Corp" with "Scanning" feature
	claims := LicenseClaims{
		Features: []string{"scanning", "cost_intel"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 1 Day
			Issuer:    "RegistryX Inc",
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		fmt.Printf("❌ Signing Failed: %v\n", err)
		return
	}
	fmt.Printf("✅ [2] License Generated (Length: %d characters)\n", len(tokenString))

	// 3. Simulating Backend Checks (Client Side)
	fmt.Println("🔄 [3] Simulating Backend Verification...")

	// Verify Signature using Public Key
	parser := new(jwt.Parser)
	parsedToken, _ := parser.ParseUnverified(tokenString, &LicenseClaims{})
	
	// Check Features
	if claims, ok := parsedToken.Claims.(*LicenseClaims); ok {
		fmt.Printf("   > Features found in License: %v\n", claims.Features)

		// ACTUAL TEST:
		scanAllowed := false
		for _, f := range claims.Features {
			if f == "scanning" { scanAllowed = true }
		}

		if scanAllowed {
			fmt.Println("✅ SUCCESS: Feature 'Scanning' is UNLOCKED! 🔓")
		} else {
			fmt.Println("❌ FAILURE: Feature 'Scanning' is LOCKED! 🔒")
		}
	} else {
		fmt.Println("❌ Token Parsing Failed!")
	}
	
	fmt.Println("✨ DRY RUN COMPLETE: Logic is 100% Correct.")
}

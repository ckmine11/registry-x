package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Run: go run scripts/generate_license.go [customer_name] [plan]
func main() {
	customer := "ACME Corp"
	plan := "ENTERPRISE"
	if len(os.Args) > 1 {
		customer = os.Args[1]
	}
	if len(os.Args) > 2 {
		plan = os.Args[2]
	}

	// Check if keys exist, else generate
	if _, err := os.Stat("private.pem"); os.IsNotExist(err) {
		fmt.Println("Generating new RSA Key Pair...")
		generateKeys()
	}

	// Load Private Key
	privKeyData, _ := os.ReadFile("private.pem")
	block, _ := pem.Decode(privKeyData)
	privKey, _ := x509.ParsePKCS1PrivateKey(block.Bytes)

	// Create Claims (The Passport Data)
	claims := jwt.MapClaims{
		"customer":   customer,
		"plan":       plan,
		"features":   []string{"scanning", "cost_intel", "sso", "audit_logs"},
		"storage_gb": 5000,
		"iss":        "RegistryX Inc",
		"iat":        time.Now().Unix(),
		"exp":        time.Now().AddDate(1, 0, 0).Unix(), // 1 Year Expiry
	}

	// Sign Token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privKey)
	if err != nil {
		fmt.Println("Error signing token:", err)
		return
	}

	fmt.Println("\n---- LICENSE KEY FOR CLIENT ----")
	fmt.Println(tokenString)
	fmt.Println("--------------------------------")
	fmt.Printf("Customer: %s\nPlan: %s\nExpiry: %s\n", customer, plan, time.Now().AddDate(1, 0, 0).Format("2006-01-02"))
}

func generateKeys() {
	// Generate Key
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	
	// Save Private Key
	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	}
	f, _ := os.Create("private.pem")
	pem.Encode(f, pemBlock)
	f.Close()

	// Save Public Key
	pubASN1, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	pemBlock = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	}
	f, _ = os.Create("public.pem")
	pem.Encode(f, pemBlock)
	f.Close()

	fmt.Println("✅ Keys Generated: private.pem (KEEP SECRET), public.pem (EMBED IN APP)")
}

package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
)

func main() {
	fmt.Println("🔐 JWT Secret Generator")

	accessSecret, err := generateSecret(32)
	if err != nil {
		fmt.Printf("Error generating access secret: %v\n", err)
		os.Exit(1)
	}

	refreshSecret, err := generateSecret(32)
	if err != nil {
		fmt.Printf("Error generating refresh secret: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Generated JWT Secrets:")
	fmt.Println("Access Token Secret:")
	fmt.Println(accessSecret)
	fmt.Println("")
	fmt.Println("Refresh Token Secret:")
	fmt.Println(refreshSecret)
	fmt.Println("")
	fmt.Println("📝 Add these to your .env file:")
	fmt.Printf("JWT_ACCESS_SECRET=%s\n", accessSecret)
	fmt.Printf("JWT_REFRESH_SECRET=%s\n", refreshSecret)
	fmt.Println("")
	fmt.Println("⚠️  Security Reminders:")
	fmt.Println("  • Never commit these secrets to version control")
	fmt.Println("  • Use different secrets for dev and production")
	fmt.Println("  • Rotate secrets regularly in production")
}

func generateSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

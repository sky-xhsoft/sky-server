package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "admin123"

	// Generate new hash
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error generating hash:", err)
		return
	}

	fmt.Println("Password:", password)
	fmt.Println("New Hash:", string(hash))

	// Verify existing hash
	existingHash := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	err = bcrypt.CompareHashAndPassword([]byte(existingHash), []byte(password))
	if err == nil {
		fmt.Println("\n✓ Existing hash is VALID for password:", password)
	} else {
		fmt.Println("\n✗ Existing hash is INVALID:", err)
	}
}

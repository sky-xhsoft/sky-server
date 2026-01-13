package main

import (
	"fmt"
	"os"

	"github.com/sky-xhsoft/sky-server/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Configuration Check ===")
	fmt.Printf("✅ Config loaded successfully\n")
	fmt.Printf("\nFile Configuration:\n")
	fmt.Printf("  UploadDir:   %s\n", cfg.File.UploadDir)
	fmt.Printf("  MaxFileSize: %d bytes\n", cfg.File.MaxFileSize)
	fmt.Printf("  MaxFileSize: %.2f GB\n", float64(cfg.File.MaxFileSize)/(1024*1024*1024))
	fmt.Printf("  AllowedExts: %v\n", cfg.File.AllowedExts)
	fmt.Println("\n=== Expected Values ===")
	fmt.Println("  MaxFileSize should be: 21474836480 bytes (20.00 GB)")

	if cfg.File.MaxFileSize == 21474836480 {
		fmt.Println("\n✅ Configuration is CORRECT!")
	} else if cfg.File.MaxFileSize == 0 {
		fmt.Println("\n❌ MaxFileSize is 0 - config not loaded or field missing!")
	} else {
		fmt.Printf("\n⚠️  MaxFileSize is %.2f GB - unexpected value!\n",
			float64(cfg.File.MaxFileSize)/(1024*1024*1024))
	}
}

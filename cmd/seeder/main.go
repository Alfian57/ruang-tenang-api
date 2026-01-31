package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/database"
)

func main() {
	// Parse command line flags
	env := flag.String("env", "development", "Environment to seed: 'production' or 'development'")
	flag.Parse()

	// Normalize environment flag
	envStr := strings.ToLower(strings.TrimSpace(*env))

	// Display banner
	log.Println("")
	log.Println("╔══════════════════════════════════════════════════════════════╗")
	log.Println("║                   RUANG TENANG SEEDER                        ║")
	log.Println("╚══════════════════════════════════════════════════════════════╝")
	log.Println("")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// Connect to database
	log.Println("📦 Connecting to database...")
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	log.Println("✅ Database connected")
	log.Println("")

	// Run appropriate seeder based on environment
	switch envStr {
	case "production", "prod":
		if err := runProductionSeeder(db); err != nil {
			log.Fatalf("❌ Production seeding failed: %v", err)
		}

	case "development", "dev":
		if err := runDevelopmentSeeder(db); err != nil {
			log.Fatalf("❌ Development seeding failed: %v", err)
		}

	default:
		log.Printf("❌ Unknown environment: %s", envStr)
		log.Println("   Valid options: 'production' (or 'prod'), 'development' (or 'dev')")
		os.Exit(1)
	}

	log.Println("")
	log.Println("╔══════════════════════════════════════════════════════════════╗")
	log.Println("║                    SEEDING COMPLETE ✅                       ║")
	log.Println("╚══════════════════════════════════════════════════════════════╝")
	log.Println("")
	log.Println("📋 Test Accounts (Development only):")
	log.Println("   Admin: admin@ruang-tenang.com / password")
	log.Println("   Member: gading@gmail.com / password")
}

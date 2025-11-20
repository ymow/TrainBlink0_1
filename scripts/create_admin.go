package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"

	"github.com/ymow/messenger_protocol_research/internal/admin"
	"github.com/ymow/messenger_protocol_research/internal/config"
	"github.com/ymow/messenger_protocol_research/internal/database"
)

// CreateAdmin creates an initial super admin user
// Usage: go run scripts/create_admin.go <email> <password> <name>
func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run scripts/create_admin.go <email> <password> <name>")
		fmt.Println("Example: go run scripts/create_admin.go admin@trainblink.com password123 \"Super Admin\"")
		os.Exit(1)
	}

	email := os.Args[1]
	password := os.Args[2]
	name := os.Args[3]

	fmt.Println("🚀 TrainBlink Admin Creation Script")
	fmt.Println("====================================")

	// Load configuration
	fmt.Println("\n📋 Loading configuration...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// Connect to database
	fmt.Println("🐘 Connecting to PostgreSQL...")
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	// Run auto migrations
	fmt.Println("🔄 Running migrations...")
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("❌ Failed to run migrations: %v", err)
	}

	// Hash password
	fmt.Println("🔐 Hashing password...")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("❌ Failed to hash password: %v", err)
	}

	// Create admin
	fmt.Println("👤 Creating admin user...")
	adminUser := &admin.Admin{
		Email:        email,
		Name:         name,
		PasswordHash: string(hashedPassword),
		Role:         "super_admin",
		Permissions:  []string{"*"}, // All permissions
		IsActive:     true,
	}

	repo := admin.NewRepository(db)
	ctx := context.Background()

	// Check if admin already exists
	existing, err := repo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		fmt.Printf("⚠️  Admin with email %s already exists\n", email)
		fmt.Println("\nExisting Admin Details:")
		fmt.Printf("   ID: %s\n", existing.ID)
		fmt.Printf("   Email: %s\n", existing.Email)
		fmt.Printf("   Name: %s\n", existing.Name)
		fmt.Printf("   Role: %s\n", existing.Role)
		fmt.Printf("   Is Active: %v\n", existing.IsActive)
		fmt.Printf("   Created At: %s\n", existing.CreatedAt)
		os.Exit(0)
	}

	// Create new admin
	if err := repo.Create(ctx, adminUser); err != nil {
		log.Fatalf("❌ Failed to create admin: %v", err)
	}

	fmt.Println("\n✅ Admin created successfully!")
	fmt.Println("\nAdmin Details:")
	fmt.Printf("   ID: %s\n", adminUser.ID)
	fmt.Printf("   Email: %s\n", adminUser.Email)
	fmt.Printf("   Name: %s\n", adminUser.Name)
	fmt.Printf("   Role: %s\n", adminUser.Role)
	fmt.Printf("   Permissions: %v\n", adminUser.Permissions)
	fmt.Printf("   Is Active: %v\n", adminUser.IsActive)

	fmt.Println("\n🎯 Next Steps:")
	fmt.Println("   1. Start the server: go run cmd/server/main_phase2.go")
	fmt.Printf("   2. Login at: POST http://localhost%s/api/v1/admin/login\n", cfg.Server.Port)
	fmt.Println("   3. Use the following credentials:")
	fmt.Printf("      Email: %s\n", email)
	fmt.Printf("      Password: %s\n", password)
}

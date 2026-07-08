package main

import (
	"fmt"
	"log"

	"github.com/casbin/casbin/v3"
	gormadapter "github.com/casbin/gorm-adapter/v3"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=172.21.0.2 user=postgres password=secret dbname=rbac_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	//db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 2. Initialize Casbin GORM Adapter using the existing DB connection
	// This automatically creates a 'casbin_rule' table in your database
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		log.Fatalf("Failed to initialize Casbin adapter: %v", err)
	}

	// 3. Create Casbin Enforcer with the model file and the database adapter
	enforcer, err := casbin.NewEnforcer("rbac_model.conf", adapter)
	if err != nil {
		log.Fatalf("Failed to create Casbin enforcer: %v", err)
	}

	// 4. Load existing policies from the database
	if err := enforcer.LoadPolicy(); err != nil {
		log.Fatalf("Failed to load policies: %v", err)
	}

	// 5. Seed Permissions & Roles (Only run once or use conditions)
	// Clear previous data for this test run demo
	enforcer.ClearPolicy()

	// Define role permissions: p = sub (role), obj (resource), act (action)
	_, _ = enforcer.AddPolicy("admin", "*", "*")
	// _, _ = enforcer.AddPolicy("alice", "reports", "write")
	// _, _ = enforcer.AddPolicy("admin", "reports", "read")
	// _, _ = enforcer.AddPolicy("member", "reports", "write")

	// Group users into roles: g = user, role
	// _, _ = enforcer.AddGroupingPolicy("alice", "admin")
	// _, _ = enforcer.AddGroupingPolicy("bob", "member")
	// _, _ = enforcer.AddGroupingPolicy("admin", "*")

	// Save the newly added policies back to the PostgreSQL database
	if err := enforcer.SavePolicy(); err != nil {
		log.Fatalf("Failed to save policies: %v", err)
	}

	// 6. Enforce Access and Check Permissions
	checkPermission(enforcer, "alice", "reports", "write") // Expected: Allowed (admin role)
	checkPermission(enforcer, "bob", "reports", "write")   // Expected: Denied (member role)
	checkPermission(enforcer, "bob", "reports", "read")    // Expected: Allowed (member role)
}

// Helper function to validate and output results
func checkPermission(e *casbin.Enforcer, sub, obj, act string) {
	allowed, err := e.Enforce(sub, obj, act)
	if err != nil {
		fmt.Printf("Error validating rule: %v\n", err)
		return
	}

	if allowed {
		fmt.Printf("✅ ACCESS GRANTED: %s can %s %s\n", sub, act, obj)
	} else {
		fmt.Printf("❌ ACCESS DENIED: %s cannot %s %s\n", sub, act, obj)
	}
}

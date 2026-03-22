package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/AnouarMohamed/StateSight/internal/storage"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("seed failed: %v", err)
	}
}

func run() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/statesight?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := storage.Open(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer pool.Close()

	result, err := storage.NewRepository(pool).SeedBaselineData(ctx)
	if err != nil {
		return err
	}

	fmt.Println("seed completed")
	fmt.Printf("workspace_id: %s\n", result.WorkspaceID)
	fmt.Printf("application_1: %s\n", result.ApplicationOneID)
	fmt.Printf("application_2: %s\n", result.ApplicationTwoID)
	fmt.Printf("incident_id: %s\n", result.IncidentID)
	return nil
}

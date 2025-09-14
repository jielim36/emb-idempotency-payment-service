package database

import (
	"context"
	"fmt"
	"log"
	"payment-service/internal/models"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitTestDatabase() (*gorm.DB, testcontainers.Container, error) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		Env:          map[string]string{"POSTGRES_PASSWORD": "testpass", "POSTGRES_DB": "testdb"},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start container: %w", err)
	}

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf(
		"host=%s user=postgres password=testpass dbname=testdb port=%s sslmode=disable TimeZone=UTC",
		host, port.Port(),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to connect to test DB: %w", err)
	}

	if err := db.AutoMigrate(&models.Payment{}, &models.Wallet{}, &models.User{}); err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to migrate test DB: %w", err)
	}

	log.Println("Test database initialized successfully")
	return db, container, nil
}

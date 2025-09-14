package database

import (
	"context"
	"fmt"
	"log"
	"payment-service/internal/models"
	"sync"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	testDB        *gorm.DB
	testContainer testcontainers.Container
	initOnce      sync.Once
	initErr       error
)

// InitTestDatabase The container will be started and the DB will be initialized only on the first call
func InitTestDatabase() (*gorm.DB, testcontainers.Container, error) {
	initOnce.Do(func() {
		testDB, testContainer, initErr = startTestDatabase()
	})
	return testDB, testContainer, initErr
}

func startTestDatabase() (*gorm.DB, testcontainers.Container, error) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image: "postgres:15",
		Env: map[string]string{
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "postgres",
		},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("5432/tcp"),
			wait.ForExec([]string{"pg_isready", "-U", "postgres"}).
				WithPollInterval(1*time.Second).
				WithStartupTimeout(60*time.Second),
		).WithStartupTimeoutDefault(120 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	dsn := fmt.Sprintf(
		"host=%s user=postgres password=testpass dbname=testdb port=%s sslmode=disable TimeZone=UTC",
		host, port.Port(),
	)

	var db *gorm.DB
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err == nil {
			break
		}

		log.Printf("Attempt %d/%d: failed to connect to test DB: %v", i+1, maxRetries, err)
		time.Sleep(2 * time.Second)

		if i == maxRetries-1 {
			_ = container.Terminate(ctx)
			return nil, nil, fmt.Errorf("failed to connect to test DB after %d attempts: %w", maxRetries, err)
		}
	}

	var pingErr error
	for i := 0; i < 5; i++ {
		sqlDB, err := db.DB()
		if err != nil {
			pingErr = err
			time.Sleep(1 * time.Second)
			continue
		}

		if err := sqlDB.Ping(); err != nil {
			pingErr = err
			time.Sleep(1 * time.Second)
			continue
		}
		pingErr = nil
		break
	}

	if pingErr != nil {
		_ = container.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to ping database: %w", pingErr)
	}

	if err := db.AutoMigrate(&models.Payment{}, &models.Wallet{}, &models.User{}); err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to migrate test DB: %w", err)
	}

	log.Println("Test database initialized successfully")
	return db, container, nil
}

// TerminateTestDatabase 在所有测试结束后关闭容器
func TerminateTestDatabase() {
	if testContainer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := testContainer.Terminate(ctx); err != nil {
			log.Printf("Failed to terminate test container: %v", err)
		} else {
			log.Println("Test database container terminated successfully")
		}
	}
}

// GetTestDB get instance
func GetTestDB() *gorm.DB {
	return testDB
}

// CleanTestData clean data before next unit test run
func CleanTestData() error {
	if testDB == nil {
		return nil
	}

	return testDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM payments").Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM wallets").Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM users").Error; err != nil {
			return err
		}
		return nil
	})
}

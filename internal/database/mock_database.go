package database

import (
	"context"
	"fmt"
	"log"
	"payment-service/internal/models"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitTestDatabase() (*gorm.DB, testcontainers.Container, error) {
	ctx := context.Background()

	// 1. 启动临时 Postgres 容器
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		Env:          map[string]string{"POSTGRES_PASSWORD": "testpass", "POSTGRES_DB": "testdb"},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start container: %w", err)
	}

	// 2. 获取容器的连接信息
	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf(
		"host=%s user=postgres password=testpass dbname=testdb port=%s sslmode=disable TimeZone=UTC",
		host, port.Port(),
	)

	// 3. 连接 GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to connect to test DB: %w", err)
	}

	// 4. 自动迁移表结构
	if err := db.AutoMigrate(&models.Payment{}, &models.Wallet{}, &models.User{}); err != nil {
		container.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to migrate test DB: %w", err)
	}

	log.Println("Test database initialized successfully")
	return db, container, nil
}

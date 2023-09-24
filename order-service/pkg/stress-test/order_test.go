package stress_test

import (
	"context"
	"fmt"
	"github.com/dreamteam/order-service/pkg/client"
	"github.com/dreamteam/order-service/pkg/db"
	"github.com/dreamteam/order-service/pkg/models"
	"github.com/dreamteam/order-service/pkg/pb"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"

	"github.com/dreamteam/order-service/pkg/service"
)

func createTestDatabase() (*gorm.DB, error) {
	dsn := "host=horton.db.elephantsql.com user=ierrxilr  password=m0tFfwPVDiaNM6kV3kqpYGxlU0oqDD-Z dbname=ierrxilr   port=5432 sslmode=disable TimeZone=Asia/Kolkata"
	testDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %v", err)
	}

	// Migrate the test database
	err = testDB.AutoMigrate(&models.Order{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate test database: %v", err)
	}

	return testDB, nil
}

func BenchmarkCreateOrder(b *testing.B) {
	testDB, err := createTestDatabase()
	assert.NoError(b, err)

	productClient := client.InitProductServiceClient("localhost:50052")
	assert.NoError(b, err)

	server := &service.Server{
		H:          db.Handler{DB: testDB},
		ProductSvc: productClient,
	}

	req := &pb.CreateOrderRequest{
		UserId:    1,
		ProductId: 1,
		Quantity:  1,
	}

	for i := 0; i < b.N; i++ {
		server.CreateOrder(context.Background(), req)
	}
}

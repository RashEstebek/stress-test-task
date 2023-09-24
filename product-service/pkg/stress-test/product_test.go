package stress_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dreamteam/product-service/pkg/models"
	"github.com/dreamteam/product-service/pkg/pb"
	"github.com/dreamteam/product-service/pkg/utils"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"

	"github.com/dreamteam/product-service/pkg/db"
	"github.com/dreamteam/product-service/pkg/services"
	"github.com/stretchr/testify/assert"
)

func createTestDatabase() (*gorm.DB, error) {
	dsn := "host=horton.db.elephantsql.com user=eonvrlzx  password=DkoRZdNcXel3ZDsDjjDcySzDIkQSYc4k dbname=eonvrlzx  port=5432 sslmode=disable TimeZone=Asia/Kolkata"
	testDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %v", err)
	}

	err = testDB.AutoMigrate(&models.Product{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate test database: %v", err)
	}

	return testDB, nil
}

func BenchmarkCreateProduct(b *testing.B) {
	testDB, err := createTestDatabase()
	assert.NoError(b, err)

	server := services.Server{
		H: db.Handler{DB: testDB},
	}

	req := &pb.CreateProductRequest{
		Name:  "Test Product",
		Stock: 20,
		Price: 100,
	}

	for i := 0; i < b.N; i++ {
		server.CreateProduct(context.Background(), req)
	}
}

func BenchmarkFindOne(b *testing.B) {
	testDB, err := createTestDatabase()
	assert.NoError(b, err)

	server := services.Server{
		H: db.Handler{DB: testDB},
		Jwt: utils.JwtWrapper{
			SecretKey:       "your-secret-key",
			ExpirationHours: 3600,
		},
	}

	product := models.Product{
		Name:  "Test Product",
		Stock: 10,
		Price: 100,
	}
	testDB.Create(&product)

	req := &pb.FindOneRequest{
		Id: product.Id,
	}

	for i := 0; i < b.N; i++ {
		server.FindOne(context.Background(), req)
	}
}

func BenchmarkDecreaseStock(b *testing.B) {
	testDB, err := createTestDatabase()
	assert.NoError(b, err)

	server := services.Server{
		H: db.Handler{DB: testDB},
		Jwt: utils.JwtWrapper{
			SecretKey:       "your-secret-key",
			ExpirationHours: 3600,
		},
	}

	product := models.Product{
		Name:  "Test Product",
		Stock: 10,
		Price: 100,
	}
	testDB.Create(&product)

	order := models.Order{
		Id:        1,
		Price:     product.Price,
		ProductId: product.Id,
		UserId:    1,
	}
	orderJSON, _ := json.Marshal(order)

	conn, _ := amqp.Dial("amqp://guest:guest@localhost:5672/")
	defer conn.Close()

	ch, _ := conn.Channel()
	defer ch.Close()
	queueName := "order_queue"
	queue, _ := ch.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)

	ch.Publish(
		"",
		queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        orderJSON,
		},
	)

	req := &pb.DecreaseStockRequest{
		Id:      product.Id,
		OrderId: order.Id,
	}

	for i := 0; i < b.N; i++ {
		server.DecreaseStock(context.Background(), req)
	}
}

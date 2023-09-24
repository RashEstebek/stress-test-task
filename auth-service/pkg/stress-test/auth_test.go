package stress_test

import (
	"context"
	"fmt"
	"github.com/dreamteam/auth-service/pkg/services"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"

	"github.com/dreamteam/auth-service/pkg/db"
	"github.com/dreamteam/auth-service/pkg/models"
	"github.com/dreamteam/auth-service/pkg/pb"
	"github.com/dreamteam/auth-service/pkg/utils"
)

func createTestDatabase() (*gorm.DB, error) {
	dsn := "host=horton.db.elephantsql.com user=moyjywkq   password=llyE2HAAR0lkdJPzqKJq7Hk5PeIf_p3t dbname=moyjywkq   port=5432 sslmode=disable TimeZone=Asia/Kolkata"
	testDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %v", err)
	}

	err = testDB.AutoMigrate(&models.User{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate test database: %v", err)
	}

	return testDB, nil
}

func BenchmarkRegister(b *testing.B) {
	testDB, err := createTestDatabase()
	assert.NoError(b, err)
	server := services.Server{
		H: db.Handler{DB: testDB},
		Jwt: utils.JwtWrapper{
			SecretKey:       "your-secret-key",
			ExpirationHours: 3600,
		},
	}
	req := &pb.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	for i := 0; i < b.N; i++ {
		server.Register(context.Background(), req)
	}
}

func BenchmarkLogin(b *testing.B) {
	testDB, err := createTestDatabase()
	assert.NoError(b, err)
	server := services.Server{
		H: db.Handler{DB: testDB},
		Jwt: utils.JwtWrapper{
			SecretKey:       "your-secret-key",
			ExpirationHours: 3600,
		},
	}
	email := "test@example.com"
	password := "password123"
	hashedPassword := utils.HashPassword(password)

	testDB.Create(&models.User{
		Email:    email,
		Password: hashedPassword,
	})

	req := &pb.LoginRequest{
		Email:    email,
		Password: password,
	}
	for i := 0; i < b.N; i++ {
		server.Login(context.Background(), req)
	}
}

func BenchmarkValidate(b *testing.B) {
	testDB, _ := createTestDatabase()

	server := services.Server{
		H: db.Handler{DB: testDB},
		Jwt: utils.JwtWrapper{
			SecretKey:       "your-secret-key",
			ExpirationHours: 3600,
		},
	}
	email := "test@example.com"
	password := "password123"
	hashedPassword := utils.HashPassword(password)
	testDB.Create(&models.User{
		Email:    email,
		Password: hashedPassword,
	})
	token, _ := server.Jwt.GenerateToken(models.User{Email: email})
	req := &pb.ValidateRequest{
		Token: token,
	}
	for i := 0; i < b.N; i++ {
		server.Validate(context.Background(), req)
	}
}

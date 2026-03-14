package service

import (
	"auth-service/internal/model"
	"auth-service/internal/repository"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Register(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	svc := NewAuthService(mockRepo)

	user := &model.User{
		Username: "testuser",
		Password: "password123",
	}

	mockRepo.On("Create", mock.MatchedBy(func(u *model.User) bool {
		return u.Username == user.Username && u.Password != "password123" // Password should be hashed
	})).Return(nil)

	err := svc.Register(user)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	svc := NewAuthService(mockRepo)

	os.Setenv("JWT_SECRET", "test_secret")

	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &model.User{
		Username: "testuser",
		Password: string(hashedPassword),
	}
	user.ID = 1

	mockRepo.On("FindByUsername", "testuser").Return(user, nil)

	req := model.LoginRequest{
		Username: "testuser",
		Password: password,
	}

	resp, err := svc.Login(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	svc := NewAuthService(mockRepo)

	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &model.User{
		Username: "testuser",
		Password: string(hashedPassword),
	}

	mockRepo.On("FindByUsername", "testuser").Return(user, nil)

	req := model.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	resp, err := svc.Login(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "invalid username or password", err.Error())
	mockRepo.AssertExpectations(t)
}

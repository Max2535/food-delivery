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

	t.Run("successful registration with default role", func(t *testing.T) {
		user := &model.User{
			Username: "testuser",
			Password: "password123",
		}

		mockRepo.On("Create", mock.MatchedBy(func(u *model.User) bool {
			return u.Username == user.Username && u.Role == model.RoleUser && u.Password != "password123"
		})).Return(nil).Once()

		err := svc.Register(user)
		assert.NoError(t, err)
		assert.Equal(t, model.RoleUser, user.Role)
	})

	t.Run("successful registration with explicit role", func(t *testing.T) {
		user := &model.User{
			Username: "rider_test",
			Password: "password123",
			Role:     model.RoleRider,
		}

		mockRepo.On("Create", mock.MatchedBy(func(u *model.User) bool {
			return u.Username == user.Username && u.Role == model.RoleRider
		})).Return(nil).Once()

		err := svc.Register(user)
		assert.NoError(t, err)
	})

	t.Run("invalid role", func(t *testing.T) {
		user := &model.User{
			Username: "invalid_role_test",
			Password: "password123",
			Role:     "invalid",
		}

		err := svc.Register(user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid role")
	})

	mockRepo.AssertExpectations(t)
}

func TestTC_AUTH_001_Login_Success(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	svc := NewAuthService(mockRepo)

	// Set up environment for RS256 signing
	os.Setenv("PRIVATE_KEY_PATH", "../../../private_key.pem")

	password := "Password123!"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &model.User{
		Username: "customer@test.com",
		Password: string(hashedPassword),
		Role:     "customer",
	}
	user.ID = 1

	mockRepo.On("FindByUsername", "customer@test.com").Return(user, nil)

	req := model.LoginRequest{
		Username: "customer@test.com",
		Password: password,
	}

	resp, err := svc.Login(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "customer", resp.Role)
	mockRepo.AssertExpectations(t)
}

func TestTC_AUTH_003_Login_InvalidCredentials(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	svc := NewAuthService(mockRepo)

	password := "Password123!"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &model.User{
		Username: "customer@test.com",
		Password: string(hashedPassword),
	}

	mockRepo.On("FindByUsername", "customer@test.com").Return(user, nil)

	req := model.LoginRequest{
		Username: "customer@test.com",
		Password: "WrongPassword",
	}

	resp, err := svc.Login(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "invalid username or password", err.Error())
	mockRepo.AssertExpectations(t)
}


// Note: TC-AUTH-002: Token Expired is typically an integration test on the middleware/gateway 
// or requires mocking time in the token generation or using a parser with specific verification.
// Since generateToken is internal and uses time.Now(), it's better verified at the handler/middleware level.


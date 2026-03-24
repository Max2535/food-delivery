package service

import (
	"auth-service/internal/model"
	"auth-service/internal/repository"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func newSvc() (AuthService, *repository.MockUserRepository, *repository.MockRefreshTokenRepository, *repository.MockGroupRepository) {
	userRepo := new(repository.MockUserRepository)
	tokenRepo := new(repository.MockRefreshTokenRepository)
	groupRepo := new(repository.MockGroupRepository)
	return NewAuthService(userRepo, tokenRepo, groupRepo), userRepo, tokenRepo, groupRepo
}

func userGroup() model.Group {
	g := model.Group{
		Name: model.GroupUser,
		Roles: []model.Role{
			{Name: model.RoleUser},
		},
	}
	g.ID = 1
	g.Roles[0].ID = 1
	return g
}

func hashRaw(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// ── Register ─────────────────────────────────────────────────────────────────

func TestAuthService_Register_Success(t *testing.T) {
	svc, userRepo, _, groupRepo := newSvc()

	group := userGroup()
	groupRepo.On("FindByName", model.GroupUser).Return(&group, nil).Once()
	userRepo.On("Create", mock.MatchedBy(func(u *model.User) bool {
		return u.Username == "alice" && u.Email == "alice@example.com" && u.GroupID == group.ID
	})).Return(nil).Once()

	user, err := svc.Register("alice", "password123", "alice@example.com")

	assert.NoError(t, err)
	assert.Equal(t, "alice", user.Username)
	assert.Equal(t, model.GroupUser, user.Group.Name)
	assert.Equal(t, []string{model.RoleUser}, user.Group.RoleNames())
	userRepo.AssertExpectations(t)
	groupRepo.AssertExpectations(t)
}

func TestAuthService_Register_PasswordIsHashed(t *testing.T) {
	svc, userRepo, _, groupRepo := newSvc()

	group := userGroup()
	groupRepo.On("FindByName", model.GroupUser).Return(&group, nil).Once()

	var captured *model.User
	userRepo.On("Create", mock.MatchedBy(func(u *model.User) bool {
		captured = u
		return true
	})).Return(nil).Once()

	svc.Register("bob", "plaintext", "bob@example.com")

	assert.NotEqual(t, "plaintext", captured.Password, "password must be hashed")
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(captured.Password), []byte("plaintext")))
}

func TestAuthService_Register_DuplicateUsername(t *testing.T) {
	svc, userRepo, _, groupRepo := newSvc()

	group := userGroup()
	groupRepo.On("FindByName", model.GroupUser).Return(&group, nil).Once()
	userRepo.On("Create", mock.Anything).Return(errors.New("duplicate key")).Once()

	_, err := svc.Register("alice", "password123", "alice@example.com")

	assert.ErrorIs(t, err, ErrUserExists)
	userRepo.AssertExpectations(t)
}

// ── Login ─────────────────────────────────────────────────────────────────────

func TestAuthService_Login_Success(t *testing.T) {
	os.Setenv("PRIVATE_KEY_PATH", "../../../private_key.pem")
	svc, userRepo, tokenRepo, _ := newSvc()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	group := userGroup()
	user := &model.User{Username: "alice", Password: string(hashed), GroupID: group.ID, Group: group}
	user.ID = 1

	userRepo.On("FindByUsername", "alice").Return(user, nil).Once()
	tokenRepo.On("Create", mock.Anything).Return(nil).Once()

	pair, err := svc.Login("alice", "secret")

	assert.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	svc, userRepo, _, _ := newSvc()

	userRepo.On("FindByUsername", "ghost").Return(nil, errors.New("not found")).Once()

	pair, err := svc.Login("ghost", "any")

	assert.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Nil(t, pair)
	userRepo.AssertExpectations(t)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	svc, userRepo, _, _ := newSvc()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	user := &model.User{Username: "alice", Password: string(hashed)}
	userRepo.On("FindByUsername", "alice").Return(user, nil).Once()

	pair, err := svc.Login("alice", "wrong")

	assert.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Nil(t, pair)
	userRepo.AssertExpectations(t)
}

// ── Refresh ───────────────────────────────────────────────────────────────────

func TestAuthService_Refresh_Success(t *testing.T) {
	os.Setenv("PRIVATE_KEY_PATH", "../../../private_key.pem")
	svc, _, tokenRepo, _ := newSvc()

	rawToken := "validrawtoken123"
	hash := hashRaw(rawToken)
	group := userGroup()
	user := model.User{Username: "alice", GroupID: group.ID, Group: group}
	user.ID = 1

	rt := &model.RefreshToken{
		TokenHash: hash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		User:      user,
	}

	tokenRepo.On("FindByTokenHash", hash).Return(rt, nil).Once()
	tokenRepo.On("DeleteByTokenHash", hash).Return(nil).Once()
	tokenRepo.On("Create", mock.Anything).Return(nil).Once()

	pair, err := svc.Refresh(rawToken)

	assert.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	tokenRepo.AssertExpectations(t)
}

func TestAuthService_Refresh_TokenNotFound(t *testing.T) {
	svc, _, tokenRepo, _ := newSvc()

	tokenRepo.On("FindByTokenHash", mock.Anything).Return(nil, errors.New("not found")).Once()

	pair, err := svc.Refresh("nonexistent")

	assert.ErrorIs(t, err, ErrInvalidToken)
	assert.Nil(t, pair)
	tokenRepo.AssertExpectations(t)
}

func TestAuthService_Refresh_ExpiredToken(t *testing.T) {
	svc, _, tokenRepo, _ := newSvc()

	rawToken := "expiredtoken"
	hash := hashRaw(rawToken)
	rt := &model.RefreshToken{
		TokenHash: hash,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	tokenRepo.On("FindByTokenHash", hash).Return(rt, nil).Once()

	pair, err := svc.Refresh(rawToken)

	assert.ErrorIs(t, err, ErrInvalidToken)
	assert.Nil(t, pair)
	tokenRepo.AssertExpectations(t)
}

// ── Logout ────────────────────────────────────────────────────────────────────

func TestAuthService_Logout_Success(t *testing.T) {
	svc, _, tokenRepo, _ := newSvc()

	rawToken := "sometoken"
	hash := hashRaw(rawToken)
	tokenRepo.On("DeleteByTokenHash", hash).Return(nil).Once()

	err := svc.Logout(rawToken)

	assert.NoError(t, err)
	tokenRepo.AssertExpectations(t)
}

// ── LogoutAll ─────────────────────────────────────────────────────────────────

func TestAuthService_LogoutAll_Success(t *testing.T) {
	svc, _, tokenRepo, _ := newSvc()

	tokenRepo.On("DeleteByUserID", uint(42)).Return(nil).Once()

	err := svc.LogoutAll(42)

	assert.NoError(t, err)
	tokenRepo.AssertExpectations(t)
}

// ── GetProfile ────────────────────────────────────────────────────────────────

func TestAuthService_GetProfile_Success(t *testing.T) {
	svc, userRepo, _, _ := newSvc()

	expected := &model.User{Username: "alice", Email: "alice@example.com"}
	expected.ID = 1
	userRepo.On("FindByID", uint(1)).Return(expected, nil).Once()

	user, err := svc.GetProfile(1)

	assert.NoError(t, err)
	assert.Equal(t, "alice", user.Username)
	userRepo.AssertExpectations(t)
}

func TestAuthService_GetProfile_NotFound(t *testing.T) {
	svc, userRepo, _, _ := newSvc()

	userRepo.On("FindByID", uint(99)).Return(nil, errors.New("not found")).Once()

	user, err := svc.GetProfile(99)

	assert.Error(t, err)
	assert.Nil(t, user)
	userRepo.AssertExpectations(t)
}

// ── ChangePassword ────────────────────────────────────────────────────────────

func TestAuthService_ChangePassword_Success(t *testing.T) {
	svc, userRepo, _, _ := newSvc()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("old"), bcrypt.DefaultCost)
	user := &model.User{Password: string(hashed)}
	user.ID = 1

	userRepo.On("FindByID", uint(1)).Return(user, nil).Once()
	userRepo.On("UpdatePassword", uint(1), mock.AnythingOfType("string")).Return(nil).Once()

	err := svc.ChangePassword(1, "old", "newpassword")

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestAuthService_ChangePassword_WrongCurrentPassword(t *testing.T) {
	svc, userRepo, _, _ := newSvc()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	user := &model.User{Password: string(hashed)}
	user.ID = 1

	userRepo.On("FindByID", uint(1)).Return(user, nil).Once()

	err := svc.ChangePassword(1, "wrong", "newpassword")

	assert.ErrorIs(t, err, ErrIncorrectPassword)
	userRepo.AssertExpectations(t)
}

func TestAuthService_ChangePassword_UserNotFound(t *testing.T) {
	svc, userRepo, _, _ := newSvc()

	userRepo.On("FindByID", uint(99)).Return(nil, errors.New("not found")).Once()

	err := svc.ChangePassword(99, "old", "new")

	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrIncorrectPassword)
	userRepo.AssertExpectations(t)
}

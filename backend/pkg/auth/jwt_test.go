package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenManager_GenerateAccessToken(t *testing.T) {
	tm := NewTokenManager("test-secret", "refresh-secret", "verify-secret", 15*time.Minute, 7*24*time.Hour, 24*time.Hour)

	userID := uuid.New()
	token, err := tm.GenerateAccessToken(userID, "test@example.com", "testuser", "free")

	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestTokenManager_ValidateAccessToken_Success(t *testing.T) {
	tm := NewTokenManager("test-secret", "refresh-secret", "verify-secret", 15*time.Minute, 7*24*time.Hour, 24*time.Hour)

	userID := uuid.New()
	token, err := tm.GenerateAccessToken(userID, "test@example.com", "testuser", "free")
	require.NoError(t, err)

	claims, err := tm.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "free", claims.SubscriptionTier)
}

func TestTokenManager_ValidateAccessToken_Expired(t *testing.T) {
	tm := NewTokenManager("test-secret", "refresh-secret", "verify-secret", -1*time.Hour, 7*24*time.Hour, 24*time.Hour)

	userID := uuid.New()
	token, err := tm.GenerateAccessToken(userID, "test@example.com", "testuser", "free")
	require.NoError(t, err)

	_, err = tm.ValidateAccessToken(token)
	assert.Error(t, err)
}

func TestTokenManager_ValidateAccessToken_InvalidSecret(t *testing.T) {
	tm1 := NewTokenManager("secret-1", "refresh-secret", "verify-secret", 15*time.Minute, 7*24*time.Hour, 24*time.Hour)
	tm2 := NewTokenManager("secret-2", "refresh-secret", "verify-secret", 15*time.Minute, 7*24*time.Hour, 24*time.Hour)

	userID := uuid.New()
	token, err := tm1.GenerateAccessToken(userID, "test@example.com", "testuser", "free")
	require.NoError(t, err)

	_, err = tm2.ValidateAccessToken(token)
	assert.Error(t, err)
}

func TestTokenManager_GenerateRefreshToken(t *testing.T) {
	tm := NewTokenManager("test-secret", "refresh-secret", "verify-secret", 15*time.Minute, 7*24*time.Hour, 24*time.Hour)

	userID := uuid.New()
	token, err := tm.GenerateRefreshToken(userID)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestTokenManager_ValidateRefreshToken_Success(t *testing.T) {
	tm := NewTokenManager("test-secret", "refresh-secret", "verify-secret", 15*time.Minute, 7*24*time.Hour, 24*time.Hour)

	userID := uuid.New()
	token, err := tm.GenerateRefreshToken(userID)
	require.NoError(t, err)

	claims, err := tm.ValidateRefreshToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), claims.Subject)
}

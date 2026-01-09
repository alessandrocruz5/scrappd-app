package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	password := "mySecurePassword123!"

	hash, err := HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestHashPassword_Empty(t *testing.T) {
	hash, err := HashPassword("")
	assert.Error(t, err)
	assert.Empty(t, hash)
}

func TestHashPassword_Unique(t *testing.T) {
	password := "mySecurePassword123!"

	hash1, err := HashPassword(password)
	require.NoError(t, err)

	hash2, err := HashPassword(password)
	require.NoError(t, err)

	// Same password should produce different hashes (due to salt)
	assert.NotEqual(t, hash1, hash2)
}

func TestCheckPassword_Success(t *testing.T) {
	password := "mySecurePassword123!"

	hash, err := HashPassword(password)
	require.NoError(t, err)

	err = CheckPassword(hash, password)
	assert.NoError(t, err)
}

func TestCheckPassword_WrongPassword(t *testing.T) {
	password := "mySecurePassword123!"
	wrongPassword := "wrongPassword"

	hash, err := HashPassword(password)
	require.NoError(t, err)

	err = CheckPassword(hash, wrongPassword)
	assert.Error(t, err)
}

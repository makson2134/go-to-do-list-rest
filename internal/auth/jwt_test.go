package auth

import (
	"testing"
	"time"
	"to-do-list/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenManager(t *testing.T) {
	secretKey := "supersecretkey"
	tokenDuration := time.Minute * 15
	tm := NewTokenManager(secretKey, tokenDuration)
	user := models.User{ID: 1, Username: "testuser"}

	t.Run("Generate and Validate OK", func(t *testing.T) {
		tokenString, err := tm.GenerateToken(user)
		require.NoError(t, err, "Token generation should not cause errors")
		require.NotEmpty(t, tokenString, "Generated token should not be empty")

		userID, err := tm.ValidateToken(tokenString)
		require.NoError(t, err, "Token validation should not cause errors")
		assert.Equal(t, user.ID, userID, "User ID from token should match the original")
	})

	t.Run("Validation errors", func(t *testing.T) {
		tmWithAnotherKey := NewTokenManager("another-secret", tokenDuration)
		tokenWithAnotherKey, err := tmWithAnotherKey.GenerateToken(user)
		require.NoError(t, err)

		tmForExpiredToken := NewTokenManager(secretKey, -time.Minute)
		expiredToken, err := tmForExpiredToken.GenerateToken(user)
		require.NoError(t, err)

		testCases := []struct {
			name  string
			token string
		}{
			{"invalid token format", "invalid.token.string"},
			{"token with different secret", tokenWithAnotherKey},
			{"expired token", expiredToken},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := tm.ValidateToken(tc.token)
				assert.Error(t, err, "Expected an error for case: %s", tc.name)
			})
		}
	})
}

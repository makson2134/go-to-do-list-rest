package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"to-do-list/internal/auth"
	"to-do-list/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware_Success(t *testing.T) {
	tokenManager := auth.NewTokenManager("test-secret", time.Minute*15)
	middleware := AuthMiddleware(tokenManager)

	testUser := models.User{ID: 123}
	token, err := tokenManager.GenerateToken(testUser)
	require.NoError(t, err)

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		userID, ok := r.Context().Value(UserIDKey).(uint)
		assert.True(t, ok, "userID should be in context")
		assert.Equal(t, testUser.ID, userID, "userID in context should match token")
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	middleware(testHandler).ServeHTTP(rr, req)

	assert.True(t, handlerCalled, "Next handler should have been called")
	assert.NotEqual(t, http.StatusUnauthorized, rr.Code, "Response status should not be Unauthorized")
}

func TestAuthMiddleware_Failure(t *testing.T) {
	tokenManager := auth.NewTokenManager("test-secret", time.Minute*15)
	middleware := AuthMiddleware(tokenManager)

	expiredTokenManager := auth.NewTokenManager("test-secret", -time.Minute)
	expiredToken, _ := expiredTokenManager.GenerateToken(models.User{ID: 1})

	testCases := []struct {
		name          string
		authHeader    string
		expectedError string
	}{
		{
			name:          "No Authorization Header",
			authHeader:    "",
			expectedError: "Authorization header is required",
		},
		{
			name:          "Invalid Header Format - No Bearer",
			authHeader:    "invalid-token",
			expectedError: "Invalid Authorization header format",
		},
		{
			name:          "Invalid Header Format - Wrong Scheme",
			authHeader:    "Basic some-token",
			expectedError: "Invalid Authorization header format",
		},
		{
			name:          "Invalid Token",
			authHeader:    "Bearer invalid-token-string",
			expectedError: "Invalid token",
		},
		{
			name:          "Expired Token",
			authHeader:    "Bearer " + expiredToken,
			expectedError: "Invalid token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handlerCalled := false
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
			})

			req := httptest.NewRequest("GET", "/", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rr := httptest.NewRecorder()

			middleware(testHandler).ServeHTTP(rr, req)

			assert.False(t, handlerCalled, "Next handler should not be called on auth failure")
			assert.Equal(t, http.StatusUnauthorized, rr.Code)
			assert.Contains(t, rr.Body.String(), tc.expectedError)
		})
	}
}

package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestGenerateTokenAndClaims(t *testing.T) {
	token, err := GenerateToken("user1", "employee", "secret")
	assert.NoError(t, err)

	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	assert.NoError(t, err)

	claims, ok := parsed.Claims.(*Claims)
	assert.True(t, ok)
	assert.Equal(t, "user1", claims.Subject)
	assert.Equal(t, "employee", claims.Role)
}

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("mypassword")
	assert.NoError(t, err)

	err = CheckPassword(hash, "mypassword")
	assert.NoError(t, err)

	err = CheckPassword(hash, "wrongpassword")
	assert.Error(t, err)
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Middleware("secret"))
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/protected", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid")
	router.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)

	validToken, err := GenerateToken("u2", "moderator", "secret")
	assert.NoError(t, err)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

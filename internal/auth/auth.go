package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"pvz-backend-service/lib/e"
)

type Claims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

func Middleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.FullPath() {
		case "/dummyLogin", "/register", "/login":
			c.Next()
			return
		}

		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
			return
		}
		tokenString := strings.TrimPrefix(auth, "Bearer ")

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
			return
		}

		c.Set("user_id", claims.Subject)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func GenerateToken(userID, role, secret string) (string, error) {
	claims := Claims{jwt.RegisteredClaims{Subject: userID, ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour))}, role}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString([]byte(secret))
	return s, e.WrapIfErr("token generation failed", err)
}

func HashPassword(p string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(h), e.WrapIfErr("hash failed", err)
}

func CheckPassword(hash, p string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(p))
	return e.WrapIfErr("password mismatch", err)
}

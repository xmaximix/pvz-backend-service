package api

import (
	"bytes"
	"encoding/json"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	api "pvz-backend-service/internal/api/types"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var _ api.ServerInterface = (*stubService)(nil)

type stubService struct {
}

func (s stubService) PostDummyLogin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"token": "tok"})
}

func (s stubService) PostRegister(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" || req.Password == "" || req.Role == "" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "Invalid input"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"token": "tok"})
}

func (s stubService) PostLogin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"token": "tok"})
}

func (s stubService) PostPvz(c *gin.Context) {
	var req struct{ City string }
	_ = c.BindJSON(&req)
	c.JSON(http.StatusCreated, gin.H{"id": "p1", "city": req.City})
}

func (s stubService) GetPvz(c *gin.Context, params api.GetPvzParams) {
	c.JSON(http.StatusOK, gin.H{"items": []string{"p1"}, "count": 1})
}

func (s stubService) PostReceptions(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"id": "r1"})
}

func (s stubService) PostProducts(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"id": "pr1"})
}

func (s stubService) PostPvzPvzIdDeleteLastProduct(c *gin.Context, pvzId openapi_types.UUID) {
	c.Status(http.StatusOK)
}

func (s stubService) PostPvzPvzIdCloseLastReception(c *gin.Context, pvzId openapi_types.UUID) {
	c.JSON(http.StatusOK, gin.H{"id": "r1"})
}

func setupRouterNoAuth() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", "moderator")
		c.Next()
	})
	RegisterHTTP(r, &stubService{})
	return r
}

func TestMain(m *testing.M) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		panic("cannot determine project root: " + err.Error())
	}
	if err := os.Chdir(root); err != nil {
		panic("cannot chdir to project root: " + err.Error())
	}
	os.Exit(m.Run())
}

func TestInvalidJSONBody(t *testing.T) {
	r := setupRouterNoAuth()

	cases := []struct {
		method, path, body string
	}{
		{"POST", "/dummyLogin", `{bad json`},
		{"POST", "/register", `{bad json`},
		{"POST", "/login", `{bad json`},
		{"POST", "/pvz", `{bad json`},
		{"POST", "/receptions", `{bad json`},
		{"POST", "/products", `{bad json`},
	}

	for _, tc := range cases {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code, "%s %s should 400", tc.method, tc.path)
	}
}

func TestMissingContentType(t *testing.T) {
	r := setupRouterNoAuth()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/pvz", bytes.NewBufferString(`{"city":"Москва"}`))
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNotFoundRoute(t *testing.T) {
	r := setupRouterNoAuth()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/not_exist", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRegisterLoginValidation(t *testing.T) {
	r := setupRouterNoAuth()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/register", bytes.NewBufferString(`{"email":"","password":"p","role":"moderator"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/login", bytes.NewBufferString(`{"email":"a@b","password":"p"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "tok", resp["token"])
}

func TestQueryParamValidationErrors(t *testing.T) {
	r := setupRouterNoAuth()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/pvz?startDate=xxx&endDate=2020-01-01T00:00:00Z&page=1&limit=10", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var errBody map[string]string
	json.Unmarshal(w.Body.Bytes(), &errBody)
	assert.Contains(t, errBody["msg"], "startDate")

	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/pvz?startDate=2020-01-01T00:00:00Z&endDate=2020-01-01T00:00:00Z&page=1&limit=bad", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	json.Unmarshal(w.Body.Bytes(), &errBody)
	assert.Contains(t, errBody["msg"], "limit")
}
func TestPathParamValidationErrors(t *testing.T) {
	r := setupRouterNoAuth()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/pvz/not-uuid/delete_last_product", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var errBody map[string]string
	json.Unmarshal(w.Body.Bytes(), &errBody)
	assert.Contains(t, errBody["msg"], "pvzId")

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/pvz/not-uuid/close_last_reception", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	json.Unmarshal(w.Body.Bytes(), &errBody)
	assert.Contains(t, errBody["msg"], "pvzId")
}

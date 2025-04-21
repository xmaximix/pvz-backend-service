package service

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"pvz-backend-service/internal/auth"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	api "pvz-backend-service/internal/api/types"
	"pvz-backend-service/internal/model"
	"pvz-backend-service/internal/repo"
)

type stubRepoSuccess struct{}

var _ repo.Repository = (*stubRepoSuccess)(nil)

func (s *stubRepoSuccess) CreateUser(_ context.Context, email, hash, role string) (model.User, error) {
	return model.User{ID: "u1", Email: email, Role: role}, nil
}
func (s *stubRepoSuccess) GetUserByEmail(_ context.Context, email string) (model.User, error) {
	hash, _ := auth.HashPassword("p")
	return model.User{ID: "u1", Email: email, PasswordHash: hash, Role: "employee"}, nil
}
func (s *stubRepoSuccess) CreatePVZ(_ context.Context, city string) (model.PVZ, error) {
	return model.PVZ{ID: "p1", City: city}, nil
}
func (s *stubRepoSuccess) ListPVZ(_ context.Context, _, _ string, _, _ int) ([]model.PVZ, error) {
	return []model.PVZ{{ID: "p1", City: "Москва"}}, nil
}
func (s *stubRepoSuccess) OpenReception(_ context.Context, pvzID string) (model.Reception, error) {
	return model.Reception{ID: "r1", PVZID: pvzID}, nil
}
func (s *stubRepoSuccess) GetOpenReception(_ context.Context, pvzID string) (model.Reception, error) {
	return model.Reception{ID: "r1", PVZID: pvzID}, nil
}
func (s *stubRepoSuccess) AddProduct(_ context.Context, recID, typ string) (model.Product, error) {
	return model.Product{ID: "pr1", ReceptionID: recID, Type: typ}, nil
}
func (s *stubRepoSuccess) DeleteLastProduct(_ context.Context, recID string) error {
	return nil
}
func (s *stubRepoSuccess) CloseReception(_ context.Context, recID string) error {
	return nil
}

type stubRepoError struct{}

var _ repo.Repository = (*stubRepoError)(nil)

func (r *stubRepoError) CreateUser(_ context.Context, _, _, _ string) (model.User, error) {
	return model.User{}, errors.New("db create user failed")
}
func (r *stubRepoError) GetUserByEmail(_ context.Context, _ string) (model.User, error) {
	return model.User{}, errors.New("db get user failed")
}
func (r *stubRepoError) CreatePVZ(_ context.Context, _ string) (model.PVZ, error) {
	return model.PVZ{}, errors.New("db create pvz failed")
}
func (r *stubRepoError) ListPVZ(_ context.Context, _, _ string, _, _ int) ([]model.PVZ, error) {
	return nil, errors.New("db list pvz failed")
}
func (r *stubRepoError) OpenReception(_ context.Context, _ string) (model.Reception, error) {
	return model.Reception{}, errors.New("db open reception failed")
}
func (r *stubRepoError) GetOpenReception(_ context.Context, _ string) (model.Reception, error) {
	return model.Reception{}, errors.New("db get open reception failed")
}
func (r *stubRepoError) AddProduct(_ context.Context, _, _ string) (model.Product, error) {
	return model.Product{}, errors.New("db add product failed")
}
func (r *stubRepoError) DeleteLastProduct(_ context.Context, _ string) error {
	return errors.New("db delete product failed")
}
func (r *stubRepoError) CloseReception(_ context.Context, _ string) error {
	return errors.New("db close reception failed")
}

func newContext(method, path, body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		c.Request.Header.Set("Content-Type", "application/json")
	}
	return c, w
}

func TestPostDummyLogin(t *testing.T) {
	svc := New(&stubRepoSuccess{}, "secret")
	c, w := newContext("POST", "/dummyLogin", `{"role":"employee"}`)
	svc.PostDummyLogin(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPostRegister(t *testing.T) {
	svc := New(&stubRepoSuccess{}, "secret")
	c, w := newContext("POST", "/register", `{"email":"a@b","password":"p","role":"employee"}`)
	svc.PostRegister(c)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestPostRegister_DBError(t *testing.T) {
	svc := New(&stubRepoError{}, "secret")
	c, w := newContext("POST", "/register", `{"email":"a@b","password":"p","role":"employee"}`)
	svc.PostRegister(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostLogin(t *testing.T) {
	svc := New(&stubRepoSuccess{}, "secret")
	c, w := newContext("POST", "/login", `{"email":"a@b","password":"p"}`)
	svc.PostLogin(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPostLogin_BadCreds(t *testing.T) {
	svc := New(&stubRepoSuccess{}, "secret")
	c, w := newContext("POST", "/login", `{"email":"a@b","password":"wrong"}`)
	svc.PostLogin(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPostPvz(t *testing.T) {
	svc := New(&stubRepoSuccess{}, "secret")
	c, w := newContext("POST", "/pvz", `{"city":"Казань"}`)
	c.Set("role", "moderator")
	svc.PostPvz(c)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestPostPvz_Forbidden(t *testing.T) {
	svc := New(&stubRepoSuccess{}, "secret")
	c, w := newContext("POST", "/pvz", `{"city":"Казань"}`)
	c.Set("role", "employee")
	svc.PostPvz(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetPvz(t *testing.T) {
	svc := New(&stubRepoSuccess{}, "secret")
	params := api.GetPvzParams{}
	c, w := newContext("GET", "/pvz", "")
	svc.GetPvz(c, params)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPostReceptions(t *testing.T) {
	svc := New(&stubRepoSuccess{}, "secret")
	id := uuid.New().String()
	c, w := newContext("POST", "/receptions", `{"pvzId":"`+id+`"}`)
	c.Set("role", "employee")
	svc.PostReceptions(c)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestPostReceptions_DBError(t *testing.T) {
	svc := New(&stubRepoError{}, "secret")
	id := uuid.New().String()
	c, w := newContext("POST", "/receptions", `{"pvzId":"`+id+`"}`)
	c.Set("role", "employee")
	svc.PostReceptions(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostProducts(t *testing.T) {
	svc := New(&stubRepoSuccess{}, "secret")
	id := uuid.New().String()
	c, w := newContext("POST", "/products", `{"pvzId":"`+id+`","type":"электроника"}`)
	c.Set("role", "employee")
	svc.PostProducts(c)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestPostProducts_DBError(t *testing.T) {
	svc := New(&stubRepoError{}, "secret")
	id := uuid.New().String()
	c, w := newContext("POST", "/products", `{"pvzId":"`+id+`","type":"электроника"}`)
	c.Set("role", "employee")
	svc.PostProducts(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostPvzPvzIdDeleteLastProduct(t *testing.T) {
	svc := New(&stubRepoSuccess{}, "secret")
	id := uuid.New()
	c, w := newContext("POST", "/pvz/"+string(id.String())+"/delete_last_product", "")
	c.Set("role", "employee")
	svc.PostPvzPvzIdDeleteLastProduct(c, id)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPostPvzPvzIdCloseLastReception(t *testing.T) {
	svc := New(&stubRepoSuccess{}, "secret")
	id := uuid.New()
	c, w := newContext("POST", "/pvz/"+string(id.String())+"/close_last_reception", "")
	c.Set("role", "employee")
	svc.PostPvzPvzIdCloseLastReception(c, id)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInvalidJSON(t *testing.T) {
	svc := New(&stubRepoSuccess{}, "secret")
	c, w := newContext("POST", "/dummyLogin", `{bad}`)
	svc.PostDummyLogin(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestForbiddenRoles(t *testing.T) {
	svc := New(&stubRepoSuccess{}, "secret")
	c, w := newContext("POST", "/pvz", `{"city":"Москва"}`)
	c.Set("role", "employee")
	svc.PostPvz(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestPostPvz_InvalidJSON(t *testing.T) {
	svc := New(&stubRepoSuccess{}, "secret")
	c, w := newContext("POST", "/pvz", `{bad json}`)
	c.Set("role", "moderator")
	svc.PostPvz(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostPvz_DBError(t *testing.T) {
	svc := New(&stubRepoError{}, "secret")
	c, w := newContext("POST", "/pvz", `{"city":"Казань"}`)
	c.Set("role", "moderator")
	svc.PostPvz(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostReceptions_InvalidJSON(t *testing.T) {
	svc := New(&stubRepoSuccess{}, "secret")
	c, w := newContext("POST", "/receptions", `{oops}`)
	c.Set("role", "employee")
	svc.PostReceptions(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostReceptions_Forbidden(t *testing.T) {
	svc := New(&stubRepoSuccess{}, "secret")
	id := uuid.New().String()
	c, w := newContext("POST", "/receptions", `{"pvzId":"`+id+`"}`)
	c.Set("role", "moderator")
	svc.PostReceptions(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestPostProducts_InvalidJSON(t *testing.T) {
	svc := New(&stubRepoSuccess{}, "secret")
	c, w := newContext("POST", "/products", `{nope}`)
	c.Set("role", "employee")
	svc.PostProducts(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostProducts_Forbidden(t *testing.T) {
	svc := New(&stubRepoSuccess{}, "secret")
	id := uuid.New().String()
	c, w := newContext("POST", "/products", `{"pvzId":"`+id+`","type":"электроника"}`)
	c.Set("role", "moderator")
	svc.PostProducts(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

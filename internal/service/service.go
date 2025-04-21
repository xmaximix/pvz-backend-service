package service

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	api "pvz-backend-service/internal/api/types"
	"pvz-backend-service/internal/auth"
	"pvz-backend-service/internal/metrics"
	"pvz-backend-service/internal/repo"
)

var _ api.ServerInterface = (*service)(nil)

type service struct {
	repo   repo.Repository
	secret string
}

func New(r repo.Repository, secret string) api.ServerInterface {
	return &service{repo: r, secret: secret}
}

func (s *service) PostDummyLogin(c *gin.Context) {
	var body struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}
	tok, _ := auth.GenerateToken(uuid.NewString(), body.Role, s.secret)
	c.JSON(http.StatusOK, gin.H{"token": tok})
}

func (s *service) PostRegister(c *gin.Context) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid registration data"})
		return
	}
	hash, _ := auth.HashPassword(body.Password)
	user, err := s.repo.CreateUser(c.Request.Context(), body.Email, hash, body.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "registration failed: " + err.Error()})
		return
	}
	tok, _ := auth.GenerateToken(user.ID, user.Role, s.secret)
	c.JSON(http.StatusCreated, gin.H{"token": tok})
}

func (s *service) PostLogin(c *gin.Context) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid login data"})
		return
	}
	user, err := s.repo.GetUserByEmail(c.Request.Context(), body.Email)
	if err != nil || auth.CheckPassword(user.PasswordHash, body.Password) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid credentials"})
		return
	}
	tok, _ := auth.GenerateToken(user.ID, user.Role, s.secret)
	c.JSON(http.StatusOK, gin.H{"token": tok})
}

func (s *service) PostPvz(c *gin.Context) {
	if c.GetString("role") != "moderator" {
		c.JSON(http.StatusForbidden, gin.H{"message": "access forbidden: moderator role required"})
		return
	}
	var body struct {
		City string `json:"city"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid PVZ data"})
		return
	}
	pvz, err := s.repo.CreatePVZ(c.Request.Context(), body.City)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to create PVZ: " + err.Error()})
		return
	}
	metrics.PvzCreated.Inc()
	c.JSON(http.StatusCreated, pvz)
}

func (s *service) GetPvz(c *gin.Context, params api.GetPvzParams) {
	start, end := "", ""
	if params.StartDate != nil {
		start = params.StartDate.Format(time.RFC3339)
	}
	if params.EndDate != nil {
		end = params.EndDate.Format(time.RFC3339)
	}
	page, limit := 1, 10
	if params.Page != nil {
		page = int(*params.Page)
	}
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	offset := (page - 1) * limit
	list, err := s.repo.ListPVZ(c.Request.Context(), start, end, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "could not list PVZs: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (s *service) PostReceptions(c *gin.Context) {
	if c.GetString("role") != "employee" {
		c.JSON(http.StatusForbidden, gin.H{"message": "access forbidden: employee role required"})
		return
	}
	var body struct {
		PVZID openapi_types.UUID `json:"pvzId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid reception data"})
		return
	}
	rec, err := s.repo.OpenReception(c.Request.Context(), body.PVZID.String())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to open reception: " + err.Error()})
		return
	}
	metrics.ReceptionCreated.Inc()
	c.JSON(http.StatusCreated, rec)
}

func (s *service) PostProducts(c *gin.Context) {
	if c.GetString("role") != "employee" {
		c.JSON(http.StatusForbidden, gin.H{"message": "access forbidden: employee role required"})
		return
	}
	var body struct {
		PVZID openapi_types.UUID `json:"pvzId"`
		Type  string             `json:"type"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid product data"})
		return
	}
	rec, err := s.repo.GetOpenReception(c.Request.Context(), body.PVZID.String())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "no open reception found"})
		return
	}
	prod, err := s.repo.AddProduct(c.Request.Context(), rec.ID, body.Type)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to add product: " + err.Error()})
		return
	}
	metrics.ProductsAdded.Inc()
	c.JSON(http.StatusCreated, prod)
}

func (s *service) PostPvzPvzIdDeleteLastProduct(c *gin.Context, pvzId openapi_types.UUID) {
	if c.GetString("role") != "employee" {
		c.JSON(http.StatusForbidden, gin.H{"message": "access forbidden: employee role required"})
		return
	}
	rec, err := s.repo.GetOpenReception(c.Request.Context(), pvzId.String())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "no open reception found"})
		return
	}
	if err := s.repo.DeleteLastProduct(c.Request.Context(), rec.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to delete last product: " + err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *service) PostPvzPvzIdCloseLastReception(c *gin.Context, pvzId openapi_types.UUID) {
	if c.GetString("role") != "employee" {
		c.JSON(http.StatusForbidden, gin.H{"message": "access forbidden: employee role required"})
		return
	}
	rec, err := s.repo.GetOpenReception(c.Request.Context(), pvzId.String())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "no open reception found"})
		return
	}
	if err := s.repo.CloseReception(c.Request.Context(), rec.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to close reception: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, rec)
}

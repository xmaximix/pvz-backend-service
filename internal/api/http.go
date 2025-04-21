package api

import (
	"context"
	"github.com/getkin/kin-openapi/openapi3filter"
	"log"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	oapimw "github.com/oapi-codegen/gin-middleware"
	api "pvz-backend-service/internal/api/types"
)

func RegisterHTTP(r *gin.Engine, svc api.ServerInterface) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("api/openapi.yaml")
	if err != nil {
		log.Fatalf("failed to load OpenAPI spec: %v", err)
	}
	if err := doc.Validate(context.Background()); err != nil {
		log.Fatalf("OpenAPI spec validation failed: %v", err)
	}

	r.Use(oapimw.OapiRequestValidatorWithOptions(doc, &oapimw.Options{
		ErrorHandler: func(c *gin.Context, message string, statusCode int) {
			c.AbortWithStatusJSON(statusCode, gin.H{"msg": message})
		},
		Options: openapi3filter.Options{
			AuthenticationFunc: func(c context.Context, input *openapi3filter.AuthenticationInput) error {
				return nil
			},
		},
	}))

	api.RegisterHandlers(r, svc)
}

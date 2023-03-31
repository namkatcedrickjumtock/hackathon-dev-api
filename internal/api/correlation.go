package api

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthorizeRequest is a middleware that authorizes http requests given based on an JWT in the Authorization header.
// Note: This middleware does NOT do authentication. The token and it's claims are assumed to be valid.
// This middleware will check that the `user_id` claim in the JWT matches the `user_id` in the request body, URL query parameters, or
// the path parameters, depending on the request type.

func GenerateCorrelationID(ctx *gin.Context) {
	correlationID := uuid.New().String()
	ctx.Set("correlationID", correlationID)
}

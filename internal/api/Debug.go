package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// AuthorizeRequest is a middleware that authorizes http requests given based on an JWT in the Authorization header.
// Note: This middleware does NOT do authentication. The token and it's claims are assumed to be valid.
// This middleware will check that the `user_id` claim in the JWT matches the `user_id` in the request body, URL query parameters, or
// the path parameters, depending on the request type.

func DebugLogs(ctx *gin.Context) {
	body := map[string]any{}

	switch ctx.Request.Method {
	case http.MethodPost:
		logger.Info().Interface("request_body", body).Str("method", "POST").Msg("request body")
	case http.MethodGet:
		logger.Info().Interface("request_body", body).Str("method", "GET").Msg("request body")
	case http.MethodPatch:
		logger.Info().Interface("request_body", body).Str("method", "PATCH").Msg("request body")
	default:
	}

	err := ctx.ShouldBindBodyWith(&body, binding.JSON)
	if err != nil {
		ctx.Next()
	}

	ctx.Next()
}

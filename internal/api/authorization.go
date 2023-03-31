package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang-jwt/jwt"
)

var PublicPathRegex = regexp.MustCompile("^/(events|current_city|webhook|health_check|categories|tickets|event)")

type AccessTokenCustomClaims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

type authHeader struct {
	AccessToken string `header:"Authorization"`
}

// AuthorizeRequest is a middleware that authorizes http requests given based on an JWT in the Authorization header.
// Note: This middleware does NOT do authentication. The token and it's claims are assumed to be valid.
// This middleware will check that the `user_id` claim in the JWT matches the `user_id` in the request body, URL query parameters, or
// the path parameters, depending on the request type.
//
//nolint:funlen
func AuthorizeRequest(ctx *gin.Context) {
	// Skip authorization for public paths
	if ok := PublicPathRegex.Match([]byte(ctx.FullPath())); ok {
		ctx.Next()

		return
	}

	if ctx.Request.Method == http.MethodOptions {
		ctx.Next()

		return
	}

	header := authHeader{}

	err := ctx.ShouldBindHeader(&header)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":  err,
			"reason": "Authorization header is missing",
		})

		ctx.Abort()

		return
	}

	accessTokenStr := strings.TrimPrefix(header.AccessToken, "Bearer ")

	claims, _ := ParseAccessToken(accessTokenStr)
	if claims.UserID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":  err,
			"reason": "user_id must be set in JWT claims",
		})
		ctx.Abort()

		return
	}

	authorized := false

	switch ctx.Request.Method {
	case http.MethodPost:
		authorized, err = authorizePos(ctx, claims.UserID)
	case http.MethodGet:
		fallthrough
	case http.MethodPatch:
		authorized, err = authorizeGetAndPatch(ctx, claims.UserID)
	default:
		//nolint:goerr113
		err = fmt.Errorf("http method not allowed")
	}

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":  err,
			"reason": "Authorization header is invalid",
		})
		ctx.Abort()

		return
	}

	if !authorized {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		ctx.Abort()

		return
	}

	ctx.Next()
}

func ParseAccessToken(tokenString string) (*AccessTokenCustomClaims, error) {
	//nolint:exhaustruct
	claims := AccessTokenCustomClaims{}

	_, _ = jwt.ParseWithClaims(tokenString, &claims, nil)

	return &claims, nil
}

func authorizePos(c *gin.Context, userID string) (bool, error) {
	var body struct {
		UserID string `json:"user_id"`
	}

	err := c.ShouldBindBodyWith(&body, binding.JSON)
	if err != nil {
		return false, err
	}

	return userID == body.UserID, nil
}

//nolint:unparam
func authorizeGetAndPatch(c *gin.Context, expectedUserID string) (bool, error) {
	userIDActual := c.Query("user_id")
	if userIDActual == "" {
		userIDActual = c.Param("user_id")
	}

	return userIDActual == expectedUserID, nil
}

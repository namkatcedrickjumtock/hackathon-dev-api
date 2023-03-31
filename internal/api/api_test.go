package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	models "github.com/Iknite-space/cliqets-api/internal/models/event"
	"github.com/Iknite-space/cliqets-api/internal/services/events/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestCreateUserAuthorixation(t *testing.T) {
	JwtUser1 := `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJ1c2VyX2lkIjoiMSJ9.OoeEPo1MHA_GdkS4tFW4RjaQ6EjNqvXQcaLbX8UIM-Q`

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockService(ctrl)
	resWriter := httptest.NewRecorder()

	mockService.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(&models.User{}, nil)

	disableAuthorization := false
	router, err := NewAPIListener(mockService, disableAuthorization, "*")
	require.Nilf(t, err, "could not set up test")

	userPayload := `{
		"user_id": "1",
		"username": "string",
		"phone_number": "string",
		"email": "string",
		"city": "string",
		"profile_image": "string",
		"country": "string"
	}`
	req := jsonPostReq("/user", userPayload, JwtUser1)

	router.ServeHTTP(resWriter, req)

	assert.Equalf(t, http.StatusOK, resWriter.Code, "wrong http status code")

	res, err := io.ReadAll(resWriter.Body)
	require.Nilf(t, err, "couldn't read response body")

	createdUser := models.User{}
	err = json.Unmarshal(res, &createdUser)
	require.Nilf(t, err, "couldn't parse response body")
}

func jsonPostReq(path string, body string, jwtStr string) *http.Request {
	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, path, strings.NewReader(body))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+jwtStr)
	req.Header.Add("Content-Length", strconv.Itoa(len(body)))

	return req
}

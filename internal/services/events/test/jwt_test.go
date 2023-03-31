package events

import (
	"context"
	"fmt"
	"testing"

	models "github.com/Iknite-space/cliqets-api/internal/models/event"
	"github.com/Iknite-space/cliqets-api/internal/persistence/mocks"
	"github.com/Iknite-space/cliqets-api/internal/services/events"
	"github.com/golang-jwt/jwt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	invalideSecret         = []byte("hezybkhtpcpyoknxfzyowazzajovhslsihkatdy")
	signinSecret           = []byte("eyJhbGciOiJIUzI1nR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c")
	ErrUnepectedSigningAlg = fmt.Errorf("unexpected error")
	ctx                    = context.Background()
)

func TestValidateJwtToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// ⭐ this structure values might change and test fails if this particular order doesn't exist in DB
	testOrder := &models.Order{
		OrderID:        "3ba343c3-e241-4b7d-96dd-1274a49fa0e7",
		UserID:         "ublwJCIf6mTUmHzg1sBqkP4giFq1",
		EventID:        "8385a0c7-3752-453a-9498-6f5dfe50056c",
		Amount:         25,
		PurchaseStatus: "SUCCESSFUL",
		Ticket: []models.OrderTicketType{
			{
				TicketType: "VIP",
				Quantity:   3,
				Price:      25,
			},
		},
	}
	purchasedTicket := models.PurchasedTicket{
		EventID:    testOrder.EventID,
		TicketType: testOrder.Ticket[0].TicketType,
		UserID:     testOrder.UserID,
		OrderID:    testOrder.OrderID,
	}

	mockRepo := mocks.NewMockRepository(ctrl)
	mockRepo.EXPECT().GetOrderByID(ctx, testOrder.OrderID).Return(testOrder, nil)
	mockRepo.EXPECT().UpdateOrderStatus(ctx, testOrder.PurchaseStatus, testOrder.OrderID).Return(&models.Order{}, nil)
	mockRepo.EXPECT().CreatePurchasedTicket(ctx, purchasedTicket).Return(&purchasedTicket, nil).Times(3)

	eventService, err := events.NewService(mockRepo, nil, string(signinSecret))
	require.Nilf(t, err, "couldn't start event service instance")
	assert.NotNil(t, eventService)

	// creating token with claims.
	createdToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"alg":    "HS256",
		"source": "CamPay",
	})

	// signing the token  with signinSecret.
	signedTokenSignature, err := createdToken.SignedString(signinSecret)
	require.Nilf(t, err, "error signing token")
	require.NotNil(t, signedTokenSignature)

	validateTokenSignature, err := eventService.TransStatus(ctx, "SUCCESSFUL", testOrder.OrderID, "25", "", "", "", signedTokenSignature)
	assert.Nilf(t, err, "Token signature doesn't match")
	assert.NotNilf(t, validateTokenSignature, "Token signature is not valid")

	// signing token with an invalide secret
	signedInvalideTokenSignature, err := createdToken.SignedString(invalideSecret)
	assert.Nilf(t, err, "error signing Token")

	validateWrongSignature, err := eventService.TransStatus(ctx, "SUCCESSFUL", testOrder.OrderID, "25", "", "", "", signedInvalideTokenSignature)
	assert.Nil(t, validateWrongSignature)
	assert.NotNilf(t, err, "invalide signin Secret")
}

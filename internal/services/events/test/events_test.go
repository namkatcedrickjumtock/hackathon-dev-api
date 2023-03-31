package events

import (
	"testing"

	models "github.com/Iknite-space/cliqets-api/internal/models/event"
	"github.com/Iknite-space/cliqets-api/internal/services/events"

	"github.com/Iknite-space/cliqets-api/internal/persistence/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEventByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	eventID := "bbd11c78-eb98-47ac-a202-bf53af7f9981"
	expectedEvent := models.Event{
		ID:     eventID,
		Ticket: []models.Ticket{},
	}

	// implement unit test
	m := mocks.NewMockRepository(ctrl)
	m.EXPECT().GetEventByID(gomock.Any(), eventID).Return(&expectedEvent, nil).Times(1)

	eventService, err := events.NewService(m, nil, "")
	require.Nilf(t, err, "got error setting up the event service")

	actualEvent, err := eventService.GetEventByID(ctx, eventID)
	require.Nilf(t, err, "got unexpected error from event service")
	assert.Equalf(t, expectedEvent.ID, actualEvent.ID, "actual eventID did not match expacted")
}

func TestFutureBookedEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockRepository(ctrl)

	eventID := "f9f1ebee-df0f-4573-999a-543d498e84a7"
	userID := "ca7ea623-5296-4d6a-bae1-8e86942b2759"
	// 3 tickets for same event.
	sameTickets := []models.PurchasedTicket{
		{
			PurchaseTicket: "",
			EventID:        eventID,
			UserID:         userID,
			EventDate:      "2028-11-28T18:23:43.800788Z",
		},
		{
			EventID:   eventID,
			UserID:    userID,
			EventDate: "2028-11-28T18:23:43.800788Z",
		},
		{
			EventID:   eventID,
			UserID:    userID,
			EventDate: "2028-11-28T18:23:43.800788Z",
		},
		{
			EventID:   eventID,
			UserID:    userID,
			EventDate: "2028-11-28T18:23:43.800788Z",
		},
	}

	mockRepo.EXPECT().GetPurchasedTickets(ctx, userID, "").Return(sameTickets, nil).Times(1)

	eventService, err := events.NewService(mockRepo, nil, "")
	require.Nilf(t, err, "couldn't start event service")
	require.NotNil(t, eventService)

	bookedEvents, err := eventService.GetBookedEvents(ctx, userID)
	assert.Nil(t, err)
	// asserts to return atleast one ticket either in future.
	assert.NotNilf(t, bookedEvents.FutureEvents, "error fetching booked events")

	// assertions for same N0 of tickets.
	assert.Equalf(t, 1, len(bookedEvents.FutureEvents), "expected 1 future ticket for thesame event")
}

func TestPastBookedEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	userID := "ca7ea623-5296-4d6a-bae1-8e86942b2759"

	// 3 tickets diff tickets for 3 diff events.
	diffTickets := []models.PurchasedTicket{
		{
			EventID:   "5de15918-c113-44c5-9d70-f51298659f8f",
			UserID:    userID,
			EventDate: "2014-11-28T18:23:43.800788Z",
		},
		{
			EventID:   "e6aded99-7b68-4ece-921a-28e48567a3d1",
			UserID:    userID,
			EventDate: "2013-11-28T18:23:43.800788Z",
		},
		{
			EventID:   "40e5b51f-fbf4-44ee-94da-84fd7bbd37c2",
			UserID:    userID,
			EventDate: "2012-11-28T18:23:43.800788Z",
		},
	}
	mockRepo.EXPECT().GetPurchasedTickets(ctx, userID, "").Return(diffTickets, nil).Times(1)

	eventService, err := events.NewService(mockRepo, nil, "")
	require.Nilf(t, err, "couldn't start event service")
	require.NotNil(t, eventService)

	bookedEvents, err := eventService.GetBookedEvents(ctx, userID)
	assert.Nil(t, err)
	// asserts to return atleast one ticket  in  past.
	assert.NotNilf(t, bookedEvents.PastEvents, "error fetching booked events expected atleast 1 event")

	assert.Equalf(t, 3, len(bookedEvents.PastEvents), "expected 3 diff events")
}

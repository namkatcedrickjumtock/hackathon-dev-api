package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	models "github.com/Iknite-space/cliqets-api/internal/models/event"
	"github.com/Iknite-space/cliqets-api/internal/persistence"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var database *sql.DB

//nolint:funlen
func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14",
		Env: []string{
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=cliqets",
			"listen_addresses = 9081",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	// db location unknown
	databaseURL := fmt.Sprintf("postgres://postgres:postgres@%s/cliqets?sslmode=disable", hostAndPort)

	log.Println("Connecting to database on url: ", databaseURL)
	// Tell docker to hard kill the container in 120 seconds
	if er := resource.Expire(120); er != nil {
		return
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 60 * time.Second

	var dbErr error
	if err = pool.Retry(func() error {
		database, dbErr = sql.Open("postgres", databaseURL)

		if dbErr != nil {
			return dbErr
		}
		dbErr = database.Ping()
		return dbErr
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s :- %s", err, dbErr)
	}

	// run migrations
	err = persistence.Migrate(database, "../../../db/migrations/", "cliqets")
	if err != nil {
		log.Fatalf("Could not run migrations:%s", err)
	}

	// Run tests
	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

var ctx = context.Background()

// user test service.
func TestUserService(t *testing.T) {
	repo, err := persistence.NewRepository(database)
	require.Nilf(t, err, "couldn't inject database instance")

	// create users
	createUser, cError := repo.CreateUser(ctx, models.User{
		UserID:      "02988fe5-4f78-47a1-a89f-931aeb8a911e",
		FirtName:    "Namkat ",
		LastName:    "Cedrick",
		PhoneNumber: "+237671738755",
		Email:       "cedrick@gmail.com",
		Gender:      "M",
		// DOB:          "200-12-21",
		CityID:       "Iowa",
		ProfileImage: "",
		Country:      "USA",
	})
	assert.Nil(t, cError)
	require.NotNilf(t, createUser, "error creating user")

	// get users
	getUser, errMsg := repo.GetUser(ctx, createUser.UserID)
	assert.NotNilf(t, getUser, "error getting user from Db")
	assert.Nil(t, errMsg)
	assert.Equalf(t, getUser, createUser, "GetUser does not match the created user")

	// update user info
	updateUser, err := repo.UpdateUserInfo(ctx, models.User{
		FirtName:    "Namkat",
		LastName:    "Junior",
		PhoneNumber: "",
		Email:       "junior@gmail.com",
		Gender:      "M",
		// DOB:          "2000-12-21",
		CityID:       "",
		ProfileImage: "",
		Country:      "",
	}, createUser.UserID)
	assert.NotNilf(t, updateUser, "error updating user")
	assert.Nilf(t, err, "update user")

	// compare exacts updated fields
	getUserUpdates, err := repo.GetUser(ctx, updateUser.UserID)
	assert.Nil(t, err)
	assert.Equalf(t, updateUser, getUserUpdates, "the updated fields doesn't match")
}

//nolintlint:funlen
func TestPurchasedTickets(t *testing.T) {
	repo, err := persistence.NewRepository(database)
	require.Nilf(t, err, "couldn't inject database instance for purchased ticket service")

	getEvents, err := repo.GetEvents(ctx, "", "", 0, 2)
	require.Nil(t, err)
	require.NotNil(t, getEvents)

	// create user
	newUser, err := repo.CreateUser(ctx, models.User{
		UserID:      "c62aa84d-d159-4766-9a23-1207a117d74a",
		FirtName:    "Akoneh",
		LastName:    "Silas",
		PhoneNumber: "",
		Email:       "silas@gmail.com",
		Gender:      "",
		// DOB:         "2000-12-02",
	})
	require.Nil(t, err)

	createOrder, err := repo.CreateOrder(ctx, models.Order{
		UserID:          "",
		EventID:         getEvents[0].ID,
		UserName:        "",
		PaymentProvider: "",
		NumberOfTickets: "",
		Amount:          0,
		OrderDate:       "",
		PurchaseStatus:  "",
		Ticket:          []models.OrderTicketType{},
	})
	require.NotNil(t, createOrder)

	assert.Nil(t, err)
	assert.NotNilf(t, newUser, "error craeting user")

	createPurchasedTicket, err := repo.CreatePurchasedTicket(ctx, models.PurchasedTicket{
		EventID:       getEvents[0].ID,
		OrderID:       createOrder.OrderID,
		Title:         "new Ticket",
		OrganizerName: "Yoder",
		TicketNum:     "32",
		HallNum:       0,
		OrderNo:       "",
		SeatNum:       0,
		TicketType:    "",
		UserID:        newUser.UserID,
		CoverImg:      "",
		Venue:         "",
		EventDate:     "",
	})
	assert.Nilf(t, err, "got error while creating purchased ticket")
	require.NotNilf(t, createPurchasedTicket, "created purchased ticket failed")

	// get purchased Tickets
	purchasedTickets, purchaseErr := repo.GetPurchasedTickets(ctx, createPurchasedTicket.UserID, createPurchasedTicket.EventID)
	assert.NotNilf(t, purchasedTickets, `purchased ticket`)
	assert.Nilf(t, purchaseErr, "got error while getting purchased ticket")
}

func TestEventService(t *testing.T) {
	repo, err := persistence.NewRepository(database)
	require.Nilf(t, err, "couldn't inject database instance for purchased ticket service")

	// get all events
	getEvents, eventErr := repo.GetEvents(ctx, "city_id", "cat_id", 0, 10)
	assert.Nilf(t, eventErr, "filter events by city and cat")
	assert.Emptyf(t, getEvents, "should be empty")

	// get events by ID
	getEventByID, IDErr := repo.GetEventByID(ctx, "city_id")
	assert.NotNilf(t, IDErr, `couldn't get all events from database`)
	assert.Nil(t, getEventByID)
}

func TestCity(t *testing.T) {
	repo, err := persistence.NewRepository(database)
	require.Nilf(t, err, "couldn't inject database instance for purchased ticket service")

	currentCity, currentErr := repo.GetUpComingEventsCities(ctx)
	assert.Nil(t, currentErr)
	assert.NotNil(t, currentCity, `fetch current city`)
}

func TestUpdateUserService(t *testing.T) {
	repo, err := persistence.NewRepository(database)
	require.Nilf(t, err, "couldn't inject database instance for purchased ticket service")

	// post purchased tickets
	updateUser, err := repo.UpdateUserInfo(ctx, models.User{}, "user_id")
	assert.NotNilf(t, err, "error updating user")
	assert.Nilf(t, updateUser, "update user")
}

func TestOrderService(t *testing.T) {
	repo, err := persistence.NewRepository(database)
	require.Nilf(t, err, "couldn't inject database instance for purchased ticket service")

	newOrder := models.Order{
		UserID:          "73387694-cad5-4bc5-8f2f-2eacd92c1e75",
		EventID:         "14a72f5a-9b5a-40f4-9021-77d252b4d683",
		UserName:        "Cedrick",
		PaymentProvider: "MTN",
		NumberOfTickets: "8",
		OrderDate:       "04-11-22",
		PhoneNumber:     "237671738755",
		PurchaseStatus:  "pending",
		Ticket: models.OrderTicketTypes{
			{
				TicketType: "Prenium",
				Quantity:   5,
				Price:      2,
			},
			{
				TicketType: "Regular",
				Quantity:   5,
				Price:      1,
			},
			{
				TicketType: "Common",
				Quantity:   5,
				Price:      1,
			},
		},
	}

	// post purchased tickets
	createOrder, err := repo.CreateOrder(ctx, newOrder)
	assert.Nil(t, err)
	assert.NotNilf(t, createOrder, "successfully created order")

	// check order id
	assert.NotEmptyf(t, createOrder.OrderID, "order id set")

	// get order by id
	getOrder, err := repo.GetOrderByID(ctx, createOrder.OrderID)

	// if bother tickets are thesame
	assert.Equalf(t, newOrder.Ticket, getOrder.Ticket, "tickets checked")

	assert.Nil(t, err)
	assert.NotNilf(t, getOrder, "order created")
	assert.Equalf(t, createOrder, getOrder, "asserts all columns are thesame")
}

func TestGetAllCategoriesService(t *testing.T) {
	repo, err := persistence.NewRepository(database)
	require.Nilf(t, err, "couldn't inject database instance for get all categories service")

	categories, err := repo.GetAllCategories(ctx)
	assert.Nil(t, err)
	assert.NotNilf(t, categories, "couldn't get all categories")
}

func TestCreateEvent(t *testing.T) {
	repo, err := persistence.NewRepository(database)
	require.Nilf(t, err, "failed dependency injection for create events service")

	// sample event
	eventSample := models.Event{
		Title:        "TRIBL NIGHTS",
		City:         "",
		CityID:       "",
		Date:         "",
		Venue:        "",
		CoverPhoto:   "",
		Ticket:       []models.Ticket{},
		Description:  "",
		Organiser:    "",
		RefundPolicy: "",
		CategoryID:   "",
	}

	event, err := repo.CreateEvent(ctx, eventSample)
	assert.Nil(t, err)
	assert.NotEmptyf(t, event.ID, "error creating event")
	assert.NotNil(t, event, "failed to create event")

	eventByID, err := repo.GetEventByID(ctx, event.ID)
	assert.Nil(t, err)
	// compare created event with fetched event
	assert.Equalf(t, event, eventByID, "error evaluating event")
}

func TestPatchEvent(t *testing.T) {
	repo, err := persistence.NewRepository(database)
	require.Nilf(t, err, "couldn't inject database instance patch event service")

	event := models.Event{
		Title:      "TRIBL NIGHTS",
		City:       "",
		Venue:      "",
		CategoryID: "",
	}

	cEvent, err := repo.CreateEvent(ctx, event)
	assert.Nil(t, err)
	assert.NotEmptyf(t, cEvent.ID, "error creating event")
	assert.NotNil(t, cEvent, "failed to create event")

	patchEventPayload := models.Event{
		Title:      "Elevation Nights",
		City:       "Buea",
		Venue:      "St Claire",
		CategoryID: "Gospel Concerts",
	}

	updatedEvent, err := repo.UpdateEvent(ctx, patchEventPayload, cEvent.ID)
	assert.Nil(t, err)
	assert.NotNil(t, updatedEvent, "couldn't update event")

	getEvent, err := repo.GetEventByID(ctx, updatedEvent.ID)
	assert.Nil(t, err)
	assert.NotNil(t, getEvent, "couldn't get updated  event")
	assert.Equalf(t, getEvent.Title, updatedEvent.Title, "updated event failed.")
}

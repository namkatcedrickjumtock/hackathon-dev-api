package api

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	models "github.com/Iknite-space/cliqets-api/internal/models/event"
	"github.com/Iknite-space/cliqets-api/internal/services/events"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/rs/zerolog"
)

var logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()

//nolint:gocyclo, funlen
func NewAPIListener(eventService events.Service, disableAuthorization bool, allowedOrigins string) (*gin.Engine, error) {
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{allowedOrigins}
	config.AllowHeaders = []string{"Origin", "authorization", "content-type"}
	config.AllowCredentials = true

	router.Use(cors.New(config))
	router.Use(GenerateCorrelationID)
	router.Use(DebugLogs)

	if !disableAuthorization {
		router.Use(AuthorizeRequest)
	}

	// route
	router.GET("/events", func(ctx *gin.Context) {
		// params
		cityID := ctx.Query("city_id")
		categoryID := ctx.Query("category_id")

		// convert params types
		//nolint
		startKey, err := strconv.ParseUint(ctx.Query("start_key"), 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "invalide start key" + err.Error(),
			})
			return
		}
		//nolint
		count, err := strconv.ParseUint(ctx.Query("count"), 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "invalide  count" + err.Error(),
			})
			return
		}

		if count == 0 {
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "count must be greater Zero!",
			})
			return
		}

		event, err := eventService.GetAllEvents(ctx, cityID, categoryID, uint(startKey), uint(count))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, event)
	})

	router.GET("/events/:id", func(ctx *gin.Context) {
		eventID := ctx.Param("id")

		event, err := eventService.GetEventByID(ctx, eventID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, event)
	})

	router.GET("/current_city", func(ctx *gin.Context) {
		//nolint
		lat, err := strconv.ParseFloat(ctx.Query("lat"), 32)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		//nolint
		lng, err := strconv.ParseFloat(ctx.Query("lng"), 32)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}
		city, err := eventService.GetCurrentCity(ctx, lng, lat)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, city)
	})

	// verifying users
	router.GET("/user/:user_id", func(ctx *gin.Context) {
		// validate req params != null
		getUser, err := eventService.GetUser(ctx, ctx.Param("user_id"))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, getUser)
	})

	// create user un signup
	router.POST("/user", func(ctx *gin.Context) {
		var newUser models.User

		if err := ctx.ShouldBindBodyWith(&newUser, binding.JSON); err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "internal server error please try again: " + err.Error(),
			})
			return
		}

		createNewUser, err := eventService.CreateUser(ctx, newUser)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, createNewUser)
	})

	// getting purchased tickets
	router.GET("/purchased_tickets", func(ctx *gin.Context) {
		if ctx.Query("user_id") == "" {
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "'user_id' is a required query parameter",
			})
			return
		}
		if ctx.Query("event_id") == "" {
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "'event_id' is a required query parameter",
			})
			return
		}
		purchasedTickets, err := eventService.GetPurchasedTickets(ctx, ctx.Query("user_id"), ctx.Query("event_id"))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, purchasedTickets)
	})

	router.GET("/purchased_tickets/:id", func(ctx *gin.Context) {
		purchasedID := ctx.Param("id")
		ticket, err := eventService.GetPurchasedTicketBYID(ctx, purchasedID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, ticket)
	})
	router.GET("/booked_events", func(ctx *gin.Context) {
		bookedEvents, err := eventService.GetBookedEvents(ctx, ctx.Query("user_id"))
		if err != nil {
			errMsg := "internal server error"
			logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msgf("%s :-> %v\n", errMsg, err)
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, bookedEvents)
	})

	router.PATCH("/user", func(ctx *gin.Context) {
		var user models.User

		if ctx.Query("user_id") == "" {
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "'user_id' is a required query parameter",
			})
			return
		}

		if err := ctx.ShouldBindBodyWith(&user, binding.JSON); err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "internal server error please try again: " + err.Error(),
			})
			return
		}

		updateInfo, err := eventService.UpdateUserInfo(ctx, user, ctx.Query("user_id"))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, updateInfo)
	})

	// create order
	router.POST("/order", func(ctx *gin.Context) {
		var order models.Order
		if err := ctx.ShouldBindBodyWith(&order, binding.JSON); err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "internal server error please try again: " + err.Error(),
			})
			return
		}
		createNewOrder, err := eventService.CreateOrder(ctx, order)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, createNewOrder)
	})
	// Get Order by id
	router.GET("/order", func(ctx *gin.Context) {
		if ctx.Query("user_id") == "" {
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "'user_id' is a required query parameter",
			})
			return
		}
		logger.Debug().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msg("UserId not empty")
		if ctx.Query("order_id") == "" {
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "'order_id' is a required query parameter",
			})
			return
		}
		logger.Debug().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msg("OrderId not empty")

		// validate req params != null
		order, err := eventService.GetOrderByID(ctx, ctx.Query("order_id"), ctx.Query("user_id"))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		logger.Debug().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msg("Successfully fetched order By Id")

		ctx.JSON(http.StatusOK, order)
	})

	// callback url
	router.GET("/webhook/campay/payments", func(ctx *gin.Context) {
		status := ctx.Query("status")
		exRef := ctx.Query("external_reference")
		amount := ctx.Query("amount")
		currency := ctx.Query("currency")
		code := ctx.Query("code")
		operaorRef := ctx.Query("operator_reference")
		signature := ctx.Query("signature")

		if signature == "" {
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "signature can't be empty",
			})
			return
		}
		logger.Debug().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msg("Signature not empty")

		orderTransaction, err := eventService.TransStatus(ctx, status, exRef, amount, currency, code, operaorRef, signature)
		if err != nil {
			logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msgf("Transaction internal server error:- %v", err)
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, orderTransaction)
	})

	// health check
	router.GET("/health_check", func(ctx *gin.Context) {
		healthChecRes := map[string]bool{"success": true}
		ctx.JSON(http.StatusOK, healthChecRes)
	})

	router.GET("/tickets/:id", func(ctx *gin.Context) {
		eventID := ctx.Param("id")

		if eventID == "" {
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "even_id can't be empty",
			})
			return
		}

		guestTickets, err := eventService.GetGuestList(ctx, eventID)
		if err != nil {
			logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msgf("Transaction internal server error:- %v", err)
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, guestTickets)
	})

	// get all events categories
	router.GET("/categories", func(ctx *gin.Context) {
		categories, err := eventService.GetAllCategories(ctx)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, categories)
	})

	router.POST("/create_event", func(ctx *gin.Context) {
		var event models.Event

		if err := ctx.ShouldBindBodyWith(&event, binding.JSON); err != nil {
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "bad request to create event" + err.Error(),
			})
			return
		}
		createdEvent, err := eventService.CreateEvent(ctx, event)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, createdEvent)
	})

	router.PATCH("/event", func(ctx *gin.Context) {
		var event models.Event

		if ctx.Query("event_id") == "" {
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "'event_id' is a required query parameter",
			})
			return
		}

		if err := ctx.ShouldBindBodyWith(&event, binding.JSON); err != nil {
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "please check event payload: " + err.Error(),
			})
			return
		}

		updatedEvent, err := eventService.UpdateEvent(ctx, event, ctx.Query("event_id"))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, updatedEvent)
	})

	return router, nil
}

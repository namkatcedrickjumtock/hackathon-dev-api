package api

import (
	"net/http"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	models "github.com/namkatcedrickjumtock/sigma-auto-api/internal/models/event"
	"github.com/namkatcedrickjumtock/sigma-auto-api/internal/services/events"
)

//nolint:gocyclo, funlen
func NewAPIListener(eventService events.Service, disableAuthorization bool, allowedOrigins string) (*gin.Engine, error) {
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{allowedOrigins}
	config.AllowHeaders = []string{"Origin", "authorization", "content-type"}
	config.AllowCredentials = true

	router.Use(cors.New(config))

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

	return router, nil
}

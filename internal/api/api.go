package api

import (
	"net/http"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	models "github.com/namkatcedrickjumtock/sigma-auto-api/internal/models/cars"
	"github.com/namkatcedrickjumtock/sigma-auto-api/internal/services/cars"
)

//nolint:gocyclo, funlen
func NewAPIListener(carService cars.Service, disableAuthorization bool, allowedOrigins string) (*gin.Engine, error) {
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{allowedOrigins}
	config.AllowHeaders = []string{"Origin", "authorization", "content-type"}
	config.AllowCredentials = true

	router.Use(cors.New(config))

	if !disableAuthorization {
		router.Use(AuthorizeRequest)
	}

	// get all cars
	router.GET("/cars", func(ctx *gin.Context) {
		// params
		cityID := ctx.Query("city_id")
		categoryID := ctx.Query("category_id")

		// convert params types
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

		cars, err := carService.GetAllCars(ctx, cityID, categoryID, uint(startKey), uint(count))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, cars)
	})
	// get car by id.
	router.GET("/cars/:id", func(ctx *gin.Context) {
		carID := ctx.Param("id")

		if carID == "" {
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "car id is a required param",
			})
			return
		}
		car, err := carService.GetCarsByID(ctx, carID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, car)
	})

	// register new car.
	router.POST("/register/car", func(ctx *gin.Context) {
		var newCar models.Cars

		if err := ctx.ShouldBindBodyWith(&newCar, binding.JSON); err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "internal server error please try again: " + err.Error(),
			})
			return
		}

		car, err := carService.RegisterCar(ctx, newCar)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, car)
	})

	router.POST("/bid", func(ctx *gin.Context) {

		var bids models.Bids

		if err := ctx.ShouldBindBodyWith(&bids, binding.JSON); err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "internal server error please try again: " + err.Error(),
			})
			return
		}

		car, err := carService.PlaceBid(ctx, bids)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, car)

	})

	router.POST("/user", func(ctx *gin.Context) {

	})

	router.GET("/user", func(ctx *gin.Context) {

	})

	router.GET("/bid/:id", func(ctx *gin.Context) {
		bidID := ctx.Param("id")

		if bidID == "" {
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "car id is a required param",
			})
			return
		}

		bid, err := carService.GetBidByID(ctx, bidID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, bid)

	})

	router.GET("/webhook/campay/payments", func(ctx *gin.Context) {

	})

	return router, nil
}

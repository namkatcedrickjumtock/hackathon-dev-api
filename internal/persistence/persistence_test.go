package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	models "github.com/namkatcedrickjumtock/sigma-auto-api/internal/models/cars"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	ctx      context.Context
	database *sql.DB
)

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14",
		Env: []string{
			"POSTGRES_PASSWORD=sigma@123",
			"POSTGRES_USER=sigma",
			"POSTGRES_DB=sigma-db",
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
	databaseUrl := fmt.Sprintf("postgres://sigma:sigma@123@%s/sigma?sslmode=disable", hostAndPort)

	log.Println("Connecting to database on url: ", databaseUrl)

	resource.Expire(120) // Tell docker to hard kill the container in 120 seconds

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 120 * time.Second
	if err = pool.Retry(func() error {
		database, err = sql.Open("postgres", databaseUrl)
		if err != nil {
			return err
		}
		return database.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// run migrations
	err = Migrate(database, "../../db/migrations/", "sigma-db")
	if err != nil {
		log.Fatalf("Could not run migrations:%s", err)
	}

	//Run tests
	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestRepositoryPg_RegisterCar(t *testing.T) {
	repo, err := NewRepository(database)
	require.Nilf(t, err, "error instantiating the persistence layer.")
	require.NotNil(t, repo)
	carData := models.Cars{
		ID:                "1",
		SellerID:          "seller123",
		CarName:           "Toyota Camry",
		DatePosted:        "2024-01-18",
		BidingPrice:       "15000",
		BidExpirationTime: "2024-02-18",
		CityID:            "city123",
		EngineType:        "V6",
		CarModel:          "Camry XLE",
		NumberOfBids:      "5",
		Mileage:           "50000",
		FuelType:          "Gasoline",
		CarphotoUrl:       "https://example.com/car.jpg",
		Category:          "Sedan",
		Description:       "Well-maintained car in excellent condition.",
	}
	newCar, err := repo.RegisterCar(ctx, carData)
	assert.NoError(t, err)
	assert.NotNilf(t, newCar, "failed to create new car")

}

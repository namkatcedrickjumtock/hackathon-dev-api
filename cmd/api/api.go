package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/ardanlabs/conf/v3"
	"github.com/joho/godotenv"
	"github.com/namkatcedrickjumtock/sigma-auto-api/internal/api"
	"github.com/namkatcedrickjumtock/sigma-auto-api/internal/persistence"
	"github.com/namkatcedrickjumtock/sigma-auto-api/internal/services/cars"
	"github.com/namkatcedrickjumtock/sigma-auto-api/internal/services/payments"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Failed: %v\n", err)
		os.Exit(1)
	}
}

//nolint:funlen
func run() error {
	var cfg struct {
		API struct {
			ListenPort string `conf:"env:LISTEN_PORT,required"`
		}
		Payments struct {
			CamPayUser     string `conf:"env:CAMPAY_USER,required"`
			BaseURL        string `conf:"env:CAMPAY_BASE_URL,required"`
			CamPayPassword string `conf:"env:CAMPAY_PASSWORD,required"`
			WebHookAppKey  string `conf:"env:WEBHOOK_APP_KEY,required"`
		}
		DB struct {
			User           string `conf:"env:DB_USER,mask,required"`
			Password       string `conf:"env:DB_PASSWORD,mask,required"`
			Host           string `conf:"env:DB_HOST,required"`
			Port           int    `conf:"env:DB_PORT,required"`
			Name           string `conf:"env:DB_NAME,required"`
			DisableTLS     bool   `conf:"env:DB_DISABLE_TLS,default:false"`
			MigrationsPath string `conf:"env:DB_MIGRATIONS_PATH,required"`
		}
		DisableAuthorization bool   `conf:"env:DISABLE_AUTHORIZATION"`
		AllowedOrigins       string `conf:"env:ALLOWED_ORIGINS,required"`
	}

	// loadDevEnv loads .env file if present
	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	help, err := conf.Parse("", &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Printf("%v\n", help)

			return nil
		}

		return fmt.Errorf("parsing config: %w", err)
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name)

	runningDBInstance, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}
	defer runningDBInstance.Close()

	err = persistence.Migrate(runningDBInstance, cfg.DB.MigrationsPath, cfg.DB.Name)
	if err != nil {
		return fmt.Errorf("failed to migrate db: %w", err)
	}

	repo, err := persistence.NewRepository(runningDBInstance)
	if err != nil {
		return err
	}

	pymentService, err := payments.NewPymentService(cfg.Payments.CamPayUser, cfg.Payments.CamPayPassword, cfg.Payments.BaseURL)
	if err != nil {
		return err
	}

	eventService, err := cars.NewService(repo, pymentService, cfg.Payments.WebHookAppKey)
	if err != nil {
		return err
	}

	//nolintlint:funlen
	listener, err := api.NewAPIListener(eventService, cfg.DisableAuthorization, cfg.AllowedOrigins)
	if err != nil {
		return err
	}

	listenAddress := fmt.Sprintf("0.0.0.0:%s", cfg.API.ListenPort)

	return listener.Run(listenAddress)
}

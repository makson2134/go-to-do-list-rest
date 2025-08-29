package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"to-do-list/internal/api/routes"
	"to-do-list/internal/auth"
	"to-do-list/internal/config"

	_ "github.com/lib/pq"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	mw "to-do-list/internal/middleware" //Во избежание конфликта имён
)

func main() {
	setupLogger()

	config, err := config.NewConfig()
	if err != nil {
		slog.Error("error loading config", slog.Any("error", err))
		os.Exit(1)
	}

	db, err := connectToDb(&config.DB)
	if err != nil {
		slog.Error("failed to connect to db", slog.Any("error", err))
		os.Exit(1)
	}
	//Сразу пинг, sql.Open не проверяет фактическое подключение, судя по дебаггингу
	err = db.Ping()
	if err != nil {
		slog.Error("database ping failed", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	tokenManager := auth.NewTokenManager(config.JWT.SecretKey, config.JWT.TTL)

	slog.Info("Starting a server", slog.String("address", config.Server.Port))

	router := routes.SetupRoutes(db, tokenManager)

	finalHandler := applyGlobalMiddleware(router)

	err = http.ListenAndServe(config.Server.Port, finalHandler)
	if err != nil {
		slog.Error("server failed to start", slog.Any("error", err))
	}
}

// TODO: вынести в конфиг уровень логгирования и json/текст
func setupLogger() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func connectToDb(cfg *config.DBCfg) (*sql.DB, error) {
	connString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(cfg.ConnMaxLifetime)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	return db, nil
}

func applyGlobalMiddleware(h http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(mw.NewCustomSlogLogger(slog.Default()))
	r.Mount("/", h)

	return r
}

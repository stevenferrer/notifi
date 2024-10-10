package main

import (
	"context"
	"database/sql"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"

	callbackh "github.com/stevenferrer/notifi/callback/handler"
	callbacksvc "github.com/stevenferrer/notifi/callback/service"
	"github.com/stevenferrer/notifi/notif"
	notifh "github.com/stevenferrer/notifi/notif/handler"
	"github.com/stevenferrer/notifi/notifihttp"
	"github.com/stevenferrer/notifi/postgres"
	"github.com/stevenferrer/notifi/postgres/migration"
	"github.com/stevenferrer/notifi/token"
	tokenh "github.com/stevenferrer/notifi/token/handler"
)

const (
	defaultDSN = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	defaultRMQ = "amqp://guest:guest@localhost:5672/"
)

func main() {
	ctx := context.Background()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Postgres
	db, err := sql.Open("postgres", envStr("DSN", defaultDSN))
	if err != nil {
		logger.Fatal().Err(err).Msg("open database")
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		logger.Fatal().Err(err).Msg("ping database")
	}

	// RabbitMQ
	conn, err := amqp.Dial(envStr("RMQ", defaultRMQ))
	if err != nil {
		logger.Fatal().Err(err).Msg("dial rabbitmq")
	}
	defer conn.Close()

	senderChan, err := conn.Channel()
	if err != nil {
		logger.Fatal().Err(err).Msg("open rmq sender channel")
	}
	defer senderChan.Close()

	workerChan, err := conn.Channel()
	if err != nil {
		logger.Fatal().Err(err).Msg("open rmq worker channel")
	}
	defer workerChan.Close()

	// migrate the database
	err = migration.Migrate(db)
	if err != nil {
		logger.Fatal().Err(err).Msg("migrate database")
	}

	// Repositories
	var (
		tokenRepo    = postgres.NewTokenRepository(db)
		callbackRepo = postgres.NewCallbackRepository(db)
		idempRepo    = postgres.NewIdempRepository(db)
		notifRepo    = postgres.NewNotifRepository(db)
	)

	// Other dependencies
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	requestSender := notifihttp.NewDefaultRequestSender().
		WithHTTPClient(httpClient)

	notifSender, err := notif.NewNotifSender(senderChan)
	if err != nil {
		logger.Fatal().Err(err).Msg("new notif sender")
	}

	// Services
	var (
		tokenSvc    = token.NewTokenService(tokenRepo)
		callbackSvc = callbacksvc.NewCallbackService(callbackRepo, tokenRepo, requestSender)
		notifSvc    = notif.NewNotifService(notifRepo, notifSender)
	)

	// Notification worker
	notifMsgProcessor := notif.NewNotifMessageProcessor(requestSender,
		callbackRepo, notifRepo, tokenRepo, idempRepo)
	notifWorker, err := notif.NewNotifWorker(workerChan, notifMsgProcessor, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("new notif worker")
	}

	// Start notification worker in the background
	go func() {
		_ = notifWorker.Start(ctx)
	}()

	// HTTP middlewares
	tokenMw := tokenh.NewTokenMw(tokenSvc)

	// HTTP handlers
	var (
		tokenHandler = tokenh.NewTokenHandler(tokenSvc, logger)
		cbHandler    = callbackh.NewCallbackHandler(callbackSvc, logger)
		notifHandler = notifh.NewNotifHandler(notifSvc, logger)
	)

	// HTTP routes
	mux := chi.NewMux()
	// mux.Use(httplog.RequestLogger(logger))

	// Test endpoint
	mux.Post("/test", func(w http.ResponseWriter, r *http.Request) {
		b, _ := ioutil.ReadAll(r.Body)
		logger.Info().Str("body", string(b)).Msg("message body")

		w.WriteHeader(http.StatusOK)
	})

	mux.Route("/", func(r chi.Router) {
		r.Use(tokenMw)
		r.Mount("/token", tokenHandler)
		r.Mount("/callbacks", cbHandler)
		r.Mount("/notifications", notifHandler)
	})

	server := &http.Server{
		Addr:           "localhost:3000",
		Handler:        mux,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// start server
	go func() {
		logger.Info().Msgf("listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil {
			logger.Fatal().Err(err).Msg("listen and serve")
		}
	}()

	// setup signal capturing
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// wait for SIGINT
	<-sigChan

	// Stop notification worker
	notifWorker.Stop()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("server shutdown")
	}
}

func envStr(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}

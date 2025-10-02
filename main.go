package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/routers"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/wranglerasynq"
	"syscall"
	"time"

	"golang.org/x/net/context"

	"github.com/go-chi/chi/v5"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func main() {
	err := logging.InitLog()
	if err != nil {
		fmt.Println("Failed to initialize log")
		os.Exit(1)
	}

	logging.LogStd(logging.LOG_LEVEL_INFO, "Initializing...")

	err = config.SetConfigs()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_FATAL, err.Error())
	}

	config.CheckRequiredEnvironmentVariables()

	err = repositories.Connect()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_FATAL, err.Error())
	}

	err = repositories.MakeMigrations()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_FATAL, err.Error())
	}

	err = repositories.InitDB()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_FATAL, err.Error())
	}

	err = repositories.ConnectToRedis()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_FATAL, fmt.Errorf("redis connection error: %w", err))
	}
	defer repositories.ShutdownAsynqClient()

	err = wranglerasynq.StartEmbeddedAsynqServer()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_FATAL, fmt.Errorf("asynq worker error: %w", err))
	}
	defer wranglerasynq.ShutDownEmbeddedAsynqServer()

	err = wranglerasynq.StartEmbeddedAsynqScheduler()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_FATAL, fmt.Errorf("asynq server error: %w", err))
	}
	defer wranglerasynq.ShutDownEmbeddedAsynqScheduler()

	logging.LogStd(logging.LOG_LEVEL_INFO, "Initializing Imagick...")
	imagick.Initialize()
	defer imagick.Terminate()

	systemSettingsRepository := repositories.NewSystemSettingsRepository(nil)
	systemSettings, err := systemSettingsRepository.GetSystemSettings()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_FATAL, err.Error())
	}

	if systemSettings.EmailPollingInterval > 0 &&
		systemSettings.ReceiptProcessingSettingsId != nil {
		err = wranglerasynq.StartEmailPolling()
		if err != nil {
			logging.LogStd(logging.LOG_LEVEL_FATAL, err.Error())
		}
	}

	pepperService := services.NewPepperService(nil)
	err = pepperService.InitPepper()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_FATAL, "Failed to initialize pepper: "+err.Error())
	}

	err = wranglerasynq.StartSystemCleanUpTasks()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_FATAL, err.Error())
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	router := routers.BuildRootRouter()
	httpServer := startHttpServer(router)

	<-stop

	wranglerasynq.ShutDownEmbeddedAsynqServer()
	wranglerasynq.ShutDownEmbeddedAsynqScheduler()
	repositories.ShutdownAsynqClient()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = httpServer.Shutdown(ctx)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_FATAL, err.Error())
	}
}

func startHttpServer(router *chi.Mux) *http.Server {
	srv := &http.Server{
		Handler:      router,
		Addr:         "0.0.0.0:8081",
		WriteTimeout: 5 * time.Minute,
		ReadTimeout:  5 * time.Minute,
	}
	logging.LogStd(logging.LOG_LEVEL_INFO, "Initialize completed")
	logging.LogStd(logging.LOG_LEVEL_INFO, "Listening on port 8081")

	go func() {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logging.LogStd(logging.LOG_LEVEL_FATAL, err.Error())
		}
	}()

	return srv
}

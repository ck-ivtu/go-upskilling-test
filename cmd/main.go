package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-test/internal/controllers"
	"go-test/internal/events"
	"go-test/internal/handlers"
	"go-test/internal/util"
	"go-test/internal/workflow"
	"go-test/repository"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
	"net/http"
	"os"
	"time"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// TODO add config
	repo, err := repository.NewRepository(&repository.ConnectionParams{
		Host:     "localhost",
		Port:     "5445",
		User:     "postgres",
		Password: "password",
		DbName:   "test",
		SSLMode:  "disable",
		Logger:   logger,
	})
	if err != nil {
		logger.Fatal("Unable to initialize repository", zap.Error(err))
	}
	defer repo.CloseConn()

	if err := repo.Migrate(); err != nil {
		logger.Fatal("Database migration failed", zap.Error(err))
	}

	c, err := createTemporalClient()
	if err != nil {
		logger.Fatal("Unable to init Temporal client ", zap.Error(err))
	}
	defer c.Close()

	producer := events.NewEventProducerConfig(logger).InitEventProducer(handlers.PackageDeliveryQueueName)
	consumer := events.NewEventConsumerConfig(logger, c, workflow.PackageDeliveryTaskQueueName).InitEventConsumer(handlers.PackageDeliveryQueueName)

	maxConcurrentActivityTaskPollers := 2
	maxConcurrentWorkflowTaskPollers := 2
	maxConcurrentActivityExecutionSize := 200
	maxConcurrentWorkflowTaskExecutionSize := 200
	workerStopTimeout := 5
	timeout := time.Duration(workerStopTimeout) * time.Second

	workerOptions := worker.Options{
		MaxConcurrentActivityTaskPollers:       maxConcurrentActivityTaskPollers,
		MaxConcurrentWorkflowTaskPollers:       maxConcurrentWorkflowTaskPollers,
		MaxConcurrentActivityExecutionSize:     maxConcurrentActivityExecutionSize,
		MaxConcurrentWorkflowTaskExecutionSize: maxConcurrentWorkflowTaskExecutionSize,
		WorkerStopTimeout:                      timeout,
	}

	w := worker.New(c, workflow.PackageDeliveryTaskQueueName, workerOptions)

	workflow.SetupWorkflow(w, repo, logger)

	ginRouter := gin.Default()
	controllers.InitializeRoutes(logger, c, ginRouter, producer)

	// TODO add config
	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", "3010"),
		Handler: ginRouter,
	}

	go func() {
		err := w.Run(worker.InterruptCh())
		if err != nil {
			logger.Fatal("Unable to start the Temporal worker", zap.Error(err))
		}
	}()

	go func() {
		logger.Info(fmt.Sprintf("Listening on port %v", 3010))

		err = server.ListenAndServe()
		if err != nil {
			logger.Fatal("Unable to start server", zap.Error(err))
		}
	}()

	go func() {
		consumer.StartConsuming()
	}()

	ops := map[string]util.Operation{
		"temporal": func(ctx context.Context) error {
			w.Stop()
			return nil
		},
		"http-server": func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
		"event-consumer": func(ctx context.Context) error {
			consumer.Dispose()
			return nil
		},
	}

	wait := util.GracefulShutdown(context.Background(), timeout, ops)

	<-wait
}

func createTemporalClient() (client.Client, error) {
	temporalClient, err := client.NewLazyClient(client.Options{
		HostPort:  "localhost:7233",
		Namespace: "default",
	})
	return temporalClient, err
}

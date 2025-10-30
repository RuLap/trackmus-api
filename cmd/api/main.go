package api

import (
	"context"
	"time"

	"github.com/RuLap/trackmus-api/internal/app/auth"
	mail_services "github.com/RuLap/trackmus-api/internal/app/mail/services"
	"github.com/RuLap/trackmus-api/internal/app/task"
	"github.com/RuLap/trackmus-api/internal/pkg/config"
	"github.com/RuLap/trackmus-api/internal/pkg/http"
	"github.com/RuLap/trackmus-api/internal/pkg/jwthelper"
	"github.com/RuLap/trackmus-api/internal/pkg/logger"
	"github.com/RuLap/trackmus-api/internal/pkg/middleware"
	"github.com/RuLap/trackmus-api/internal/pkg/rabbitmq"
	"github.com/RuLap/trackmus-api/internal/pkg/redis"
	"github.com/RuLap/trackmus-api/internal/pkg/server"
	postgres "github.com/RuLap/trackmus-api/internal/pkg/storage"
	"github.com/RuLap/trackmus-api/internal/pkg/storage/minio"
	validation "github.com/RuLap/trackmus-api/internal/pkg/validator"
	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	//Config-----------------------------------------------------------------------------------------------------------

	cfg := config.MustLoad()

	logger := logger.New(logger.Config{
		Level:   cfg.Env,
		LokiURL: cfg.Log.LokiURL,
		Labels:  cfg.Log.LokiLabels,
	})

	validation.Init()

	//Additional Services-----------------------------------------------------------------------------------------------
	redisClient, err := redis.NewClient(cfg.Redis, logger)
	if err != nil {
		logger.Error("failed to connect to Redis", "error", err)
		return
	}
	logger.Info("init redis client successfully")

	redisService := redis.NewService(redisClient)
	logger.Info("init redis service successfully")

	mqClient, err := rabbitmq.NewClient(&cfg.RabbitMQ, logger)
	if err != nil {
		logger.Error("failed to connect to RabbitMQ", "error", err)
		return
	}
	defer mqClient.Close()
	logger.Info("init rabbitmq client successfully")

	mqService := rabbitmq.NewService(mqClient)
	logger.Info("init rabbitmq service successfully")

	storage, err := postgres.InitDB(cfg.PostgresConnString)
	if err != nil {
		logger.Error("failed to initialize database", "error", err)
		return
	}

	jwtHelper, err := jwthelper.NewJwtHelper(cfg.JWT.Secret)
	if err != nil {
		logger.Error("failed to create JWT helper", "error", err)
		return
	}

	minioClient, err := minio.New(&cfg.MinioConfig)
	if err != nil {
		logger.Error("failed to init MinIO: %w", err)
	}

	minioService := minio.NewService(minioClient)

	//Modules----------------------------------------------------------------------------------------------------------

	authModule := auth.NewModule(logger, storage.Database(), jwtHelper, &cfg.GoogleOAuth, redisService, mqService)
	taskModule := task.NewModule(logger, storage.Database(), minioService)

	var mailService *mail_services.MailService
	if mqService != nil {
		mailService = mail_services.NewMailService(
			logger,
			mqService,
			&cfg.SMTP,
		)

		go func() {
			logger.Info("starting mail service consumer")
			if err := mailService.StartConsumer(context.Background()); err != nil {
				logger.Error("mail service consumer failed", "error", err)
			}
		}()
	} else {
		logger.Warn("mail service not started - RabbitMQ not available")
	}
	logger.Info("Init mail service successfully")

	//Router-----------------------------------------------------------------------------------------------------------

	router := chi.NewRouter()

	router.Use(chi_middleware.RequestID)
	router.Use(chi_middleware.RealIP)
	router.Use(http.RequestLogger(logger))
	router.Use(http.Recover(logger))
	router.Use(chi_middleware.Timeout(60 * time.Second))

	router.Route("/auth", func(r chi.Router) {
		r.Post("/register", authModule.Handler.Register)
		r.Post("/login", authModule.Handler.Login)
		r.Post("/google", authModule.Handler.GoogleAuth)
		r.Get("/google/url", authModule.Handler.GoogleAuthURL)
		r.Post("/refresh", authModule.Handler.RefreshTokens)

		r.Get("/google/callback", authModule.Handler.GoogleCallback)

		r.Route("/email", func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(jwtHelper))

			r.Post("/send-confirmation", authModule.Handler.SendConfirmationLink)
			r.Post("/confirm", authModule.Handler.ConfirmEmail)
		})

		r.With(middleware.AuthMiddleware(jwtHelper)).Post("/logout", authModule.Handler.Logout)
	})

	router.Route("/tasks", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtHelper))

		r.Get("/active", taskModule.Handler.GetActiveTasks)
		r.Get("/completed", taskModule.Handler.GetCompletedTasks)
		r.Get("/{id}", taskModule.Handler.GetTaskByID)
		r.Post("/", taskModule.Handler.CreateTask)
		r.Put("/", taskModule.Handler.UpdateTask)
		r.Put("/{id}/complete", taskModule.Handler.CompleteTask)
		r.Get("{task_id}/media/upload-url", taskModule.Handler.GetMediaUploadURL)
	})

	router.Route("/sessions", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtHelper))

		r.Get("/{id}", taskModule.Handler.GetSessionByID)
		r.Post("/", taskModule.Handler.CreateSession)
	})

	router.Route("/media", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtHelper))

		r.Post("/{id}", taskModule.Handler.ConfirmMediaUpload)
		r.Delete("/", taskModule.Handler.RemoveMedia)
	})

	router.Route("/links", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtHelper))

		r.Post("/", taskModule.Handler.CreateLink)
		r.Delete("/", taskModule.Handler.RemoveLink)
	})

	//Server-----------------------------------------------------------------------------------------------------------

	server.New(router, cfg.HTTPServer)
	logger.Info("starting", "address", cfg.HTTPServer.Address)
}

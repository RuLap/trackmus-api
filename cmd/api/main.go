package api

import (
	"context"
	"log/slog"
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
	"github.com/RuLap/trackmus-api/internal/pkg/server"
	postgres "github.com/RuLap/trackmus-api/internal/pkg/storage"
	validation "github.com/RuLap/trackmus-api/internal/pkg/validator"
	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
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

	redisClient := initRedis(logger, &cfg.Redis)
	logger.Info("init redis successfully")

	rabbitmqClient := initRabbitMQ(logger, &cfg.RabbitMQ)
	logger.Info("init rabbitmq successfully")

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

	//Modules----------------------------------------------------------------------------------------------------------

	authModule := auth.NewModule(logger, storage.Database(), jwtHelper, &cfg.GoogleOAuth, redisClient, rabbitmqClient)
	taskModule := task.NewModule(logger, storage.Database(), redisClient, rabbitmqClient)

	var mailService *mail_services.MailService
	if rabbitmqClient != nil {
		mailService = mail_services.NewMailService(
			logger,
			rabbitmqClient,
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
		r.Put("{id}/complete", taskModule.Handler.CompleteTask)
	})

	router.Route("/sessions", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtHelper))

		r.Get("/{id}", taskModule.Handler.GetSessionByID)
		r.Post("/", taskModule.Handler.CreateSession)
	})

	router.Route("/media", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtHelper))

		r.Post("/", taskModule.Handler.UploadMedia)
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

func initRedis(logger *slog.Logger, cfg *config.RedisConfig) *redis.Client {
	logger.Info("starting redis", "address", cfg.Address)
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Error("failed to connect to redis", "error", err)
	}

	logger.Info("redis connected successfully")
	return rdb
}

func initRabbitMQ(logger *slog.Logger, cfg *config.RabbitMQConfig) *rabbitmq.Client {
	var rabbitmqClient *rabbitmq.Client
	var err error

	if cfg.URL != "" {
		rabbitmqClient, err = rabbitmq.NewClient(cfg.URL, logger)
		if err != nil {
			logger.Error("failed to connect to RabbitMQ",
				"error", err,
				"url", cfg.URL,
			)
		} else {
			logger.Info("successfully connected to RabbitMQ")
		}
	} else {
		logger.Warn("RabbitMQ URL not configured - email notifications will be disabled")
	}

	return rabbitmqClient
}

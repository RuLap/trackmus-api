package auth

import (
	"log/slog"

	"github.com/RuLap/trackmus-api/internal/pkg/config"
	"github.com/RuLap/trackmus-api/internal/pkg/jwthelper"
	"github.com/RuLap/trackmus-api/internal/pkg/rabbitmq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Module struct {
	Repo    Repository
	Service Service
	Handler Handler
}

func NewModule(
	log *slog.Logger,
	pool *pgxpool.Pool,
	jwtHelper *jwthelper.JWTHelper,
	googleCfg *config.GoogleOAuth,
	redis *redis.Client,
	rabbitmq *rabbitmq.Client,
) *Module {
	googleConfig := &GoogleOAuthConfig{
		ClientID:     googleCfg.ClientID,
		ClientSecret: googleCfg.ClientSecret,
		RedirectURL:  googleCfg.RedirectURL,
	}

	repo := NewRepository(pool)
	service := NewService(log, jwtHelper, googleConfig, redis, rabbitmq, repo)
	handler := NewHandler(service)

	return &Module{
		Repo:    repo,
		Service: service,
		Handler: *handler,
	}
}

package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/FrangipaneTeam/crown/pkg/config"
	"github.com/gregjones/httpcache"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/rcrowley/go-metrics"
	"github.com/rs/zerolog"
)

func main() {
	config, err := config.ReadConfig("config.yaml")
	if err != nil {
		panic(err)
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	zerolog.DefaultContextLogger = &logger

	metricsRegistry := metrics.DefaultRegistry

	cc, err := githubapp.NewDefaultCachingClientCreator(
		config.Github,
		githubapp.WithClientUserAgent("crown/1.0.0"),
		githubapp.WithClientTimeout(3*time.Second),
		githubapp.WithClientCaching(false, func() httpcache.Cache { return httpcache.NewMemoryCache() }),
		githubapp.WithClientMiddleware(
			githubapp.ClientMetrics(metricsRegistry),
		),
	)
	if err != nil {
		panic(err)
	}

	prCommentHandler := &PRCommentHandler{
		ClientCreator: cc,
		preamble:      config.AppConfig.PullRequestPreamble,
	}

	webhookHandler := githubapp.NewDefaultEventDispatcher(config.Github, prCommentHandler)

	http.Handle(githubapp.DefaultWebhookRoute, webhookHandler)

	addr := fmt.Sprintf("%s:%d", config.Server.Address, config.Server.Port)
	logger.Info().Msgf("Starting server on %s...", addr)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}

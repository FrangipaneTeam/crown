package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gregjones/httpcache"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/rcrowley/go-metrics"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/FrangipaneTeam/crown/handlers"
	"github.com/FrangipaneTeam/crown/pkg/config"
	"github.com/FrangipaneTeam/crown/pkg/db"
	"github.com/FrangipaneTeam/crown/pkg/tracker"
)

// var Db *db.DB

func main() {
	config, err := config.ReadConfig("config.yaml")
	if err != nil {
		panic(err)
	}

	// SRC https://github.com/rs/zerolog#add-file-and-line-number-to-log
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		return file + ":" + strconv.Itoa(line)
	}
	var logger zerolog.Logger

	if config.Log.Human {
		logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).With().Caller().Logger()
	} else {
		logger = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	}
	zerolog.DefaultContextLogger = &logger

	level, err := zerolog.ParseLevel(config.Log.Level)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to parse log level")
	}

	zerolog.SetGlobalLevel(level)

	err = db.NewDB(config.DB.Path)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to open database")
	}

	tracker.Init(logger)
	go tracker.Watch()

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
		logger.Fatal().Err(err).Msg("Failed to create client githubApp")
	}

	webhookHandler := githubapp.NewEventDispatcher(
		[]githubapp.EventHandler{
			&handlers.PullRequestHandler{ClientCreator: cc},
			&handlers.IssueCommentHandler{ClientCreator: cc},
			// &handlers.IssuesHandler{ClientCreator: cc},
		},
		config.Github.App.WebhookSecret,
		githubapp.WithScheduler(
			githubapp.QueueAsyncScheduler(
				100, 10,
			),
		),
	)

	http.Handle(githubapp.DefaultWebhookRoute, webhookHandler)

	addr := fmt.Sprintf("%s:%d", config.Server.Address, config.Server.Port)
	logger.Info().Msgf("Starting server on %s...", addr)

	server := &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: 3 * time.Second,
	}

	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}

	db.DataBase.Close()
}

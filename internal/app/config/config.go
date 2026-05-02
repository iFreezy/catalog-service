package config

import (
	"errors"
	"io"
	"os"
	"time"

	"github.com/iFreezy/catalog-service/internal/app/config/section"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type (
	Config struct {
		Repository section.Repository
		WebServer  section.WebServer
		Monitor    section.Monitor
	}

	LoadArgs struct {
		Output          io.Writer `json:"-"`
		EnableSimpleLog bool
		SkipConfig      bool
	}
)

var Root Config

func createLogger(level zerolog.Level, output io.Writer) zerolog.Logger {
	return zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Logger()
}

func Load(args LoadArgs) {

	zerolog.TimestampFieldName = "timestamp"
	zerolog.MessageFieldName = "msg"
	zerolog.TimeFieldFormat = time.RFC3339

	if args.Output == nil {
		args.Output = os.Stdout
	}

	if args.EnableSimpleLog {
		args.Output = zerolog.ConsoleWriter{Out: args.Output, TimeFormat: time.RFC3339}
	}

	log.Logger = createLogger(zerolog.DebugLevel, args.Output)
	log.Debug().Msg("Logger initialized with Debug level")

	if args.SkipConfig {
		log.Debug().Msg("Config loading skipped")
		return
	}

	if err := godotenv.Load(); err != nil {
		var pathErr *os.PathError
		if !errors.As(err, &pathErr) {
			log.Fatal().Err(err).Msg("load .env")
		}
	}

	if err := envconfig.Process("REPOSITORY", &Root.Repository); err != nil {
		log.Fatal().Err(err).Msg("parse REPOSITORY config")
	}
	if err := envconfig.Process("WEB_SERVER", &Root.WebServer); err != nil {
		log.Fatal().Err(err).Msg("parse WEB_SERVER config")
	}
	if err := envconfig.Process("APP", &Root.Monitor); err != nil {
		log.Fatal().Err(err).Msg("parse APP config")
	}

	level, err := zerolog.ParseLevel(Root.Monitor.LogLevel)
	if err != nil {
		log.Fatal().Err(err).Str("log_level", Root.Monitor.LogLevel).Msg("parse log level")
	}

	log.Logger = createLogger(level, args.Output)
	log.Info().
		Str("level", level.String()).
		Str("env", Root.Monitor.Environment).
		Msg("Logger reinitialized with config level")
}

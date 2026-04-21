package service_wait

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/urfave/cli/v3"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var LogLevel = new(slog.LevelVar)
var Log = func() *slog.Logger {
	LogLevel.Set(slog.LevelInfo)
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: LogLevel,
	}))
}()

const EnvBase = "SERVICE_WAIT"

func GetFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "timeout",
			Aliases: []string{"t"},
			Sources: cli.EnvVars(EnvBase + "_TIMEOUT"),
			Value:   "30s",
		},
		&cli.StringFlag{
			Name:    "interval",
			Aliases: []string{"i"},
			Sources: cli.EnvVars(EnvBase + "_INTERVAL"),
			Value:   "30s",
		},
		&cli.StringFlag{
			Name:    "url",
			Aliases: []string{"u"},
			Sources: cli.EnvVars(EnvBase + "_URL"),
			Usage:   "Probe HTTP Endpoint",
		},
		&cli.StringSliceFlag{
			Name:    "urls",
			Aliases: []string{"U"},
			Sources: cli.EnvVars(EnvBase + "_URLS"),
			Usage:   "Probe multiple HTTP Endpoints (comma separated)",
		},
		&cli.StringFlag{
			Name:    "psql-host",
			Aliases: []string{"H"},
			Sources: cli.EnvVars(EnvBase+"_PSQL_HOST", "PGHOST"),
		},
		&cli.IntFlag{
			Name:    "psql-port",
			Aliases: []string{"p"},
			Sources: cli.EnvVars(EnvBase+"_PSQL_PORT", "PGPORT"),
			Value:   5432,
		},
		&cli.StringFlag{
			Name:    "psql-user",
			Aliases: []string{"u"},
			Sources: cli.EnvVars(EnvBase+"_PSQL_USER", "PGUSER"),
		},
		&cli.StringFlag{
			Name:    "psql-password",
			Aliases: []string{"P"},
			Sources: cli.EnvVars(EnvBase+"_PSQL_PASSWORD", "PGPASSWORD"),
		},
		&cli.StringFlag{
			Name:    "psql-database",
			Aliases: []string{"d"},
			Sources: cli.EnvVars(EnvBase+"_PSQL_DATABASE", "PGDATABASE"),
		},
		&cli.StringFlag{
			Name:    "psql-sslmode",
			Aliases: []string{"s"},
			Sources: cli.EnvVars(EnvBase+"_PSQL_SSLMODE", "PGSSLMODE"),
			Value:   "disable",
		},
		&cli.StringFlag{
			Name:    "psql-dsn",
			Sources: cli.EnvVars(EnvBase + "_PSQL_DSN"),
		},
		&cli.StringFlag{
			Name:    "mongo-host",
			Sources: cli.EnvVars(EnvBase + "_MONGO_HOST"),
		},
		&cli.IntFlag{
			Name:    "mongo-port",
			Sources: cli.EnvVars(EnvBase + "_MONGO_PORT"),
			Value:   27017,
		},
		&cli.StringFlag{
			Name:    "mongo-user",
			Sources: cli.EnvVars(EnvBase + "_MONGO_USER"),
		},
		&cli.StringFlag{
			Name:    "mongo-password",
			Sources: cli.EnvVars(EnvBase + "_MONGO_PASSWORD"),
		},
		&cli.StringFlag{
			Name:    "mongo-database",
			Sources: cli.EnvVars(EnvBase + "_MONGO_DATABASE"),
		},
		&cli.StringFlag{
			Name:    "mongo-auth-source",
			Sources: cli.EnvVars(EnvBase + "_MONGO_AUTH_SOURCE"),
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"V"},
			Sources: cli.EnvVars(EnvBase + "_DEBUG"),
		},
	}
}

type Config struct {
	Timeout         string
	Interval        string
	URL             string
	URLS            []string
	PsqlHost        string
	PsqlPort        int
	PsqlUser        string
	PsqlPassword    string
	PsqlDatabase    string
	PsqlSSLMode     string
	PsqlDsn         string
	MongoHost       string
	MongoPort       int
	MongoUser       string
	MongoPassword   string
	MongoDatabase   string
	MongoAuthSource string
	Debug           bool
}

func New(cmd *cli.Command) (*Config, error) {
	config := &Config{
		Timeout:         cmd.String("timeout"),
		Interval:        cmd.String("interval"),
		URL:             cmd.String("url"),
		URLS:            cmd.StringSlice("urls"),
		PsqlHost:        cmd.String("psql-host"),
		PsqlPort:        cmd.Int("psql-port"),
		PsqlUser:        cmd.String("psql-user"),
		PsqlPassword:    cmd.String("psql-password"),
		PsqlDatabase:    cmd.String("psql-database"),
		PsqlSSLMode:     cmd.String("psql-sslmode"),
		PsqlDsn:         cmd.String("psql-dsn"),
		MongoHost:       cmd.String("mongo-host"),
		MongoPort:       cmd.Int("mongo-port"),
		MongoUser:       cmd.String("mongo-user"),
		MongoPassword:   cmd.String("mongo-password"),
		MongoDatabase:   cmd.String("mongo-database"),
		MongoAuthSource: cmd.String("mongo-auth-source"),
		Debug:           cmd.Bool("verbose"),
	}
	err := validateConfig(config)
	if err != nil {
		return nil, err
	}
	if config.Debug {
		LogLevel.Set(slog.LevelDebug)
		Log.Debug("Enabling debug logging")
	}
	return config, nil
}

func ProbeHttpEndpoint(ctx context.Context, config *Config) error {
	timeoutDuration, intervalDuration, err := parseDurations(config.Timeout, config.Interval)
	if err != nil {
		return err
	}

	var urls []string

	if config.URL != "" {
		urls = []string{config.URL}
	}
	if config.URLS != nil && len(config.URLS) > 0 {
		urls = append(urls, config.URLS...)
	}

	for _, _url := range urls {
		for {
			Log.Info(fmt.Sprintf("Probing HTTP endpoint url=%s timeout=%s interval=%s", _url, config.Timeout, config.Interval))
			client := &http.Client{Timeout: timeoutDuration}
			resp, err := client.Get(_url)

			if err != nil {
				Log.Debug(err.Error())
			}
			if err == nil && (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent) {
				resp.Body.Close()
				break
			}
			select {
			case <-ctx.Done():
				return fmt.Errorf("cancled while waiting for %s: %w", config.URL, ctx.Err())
			default:
				time.Sleep(intervalDuration)
			}
		}
	}
	return nil
}

func ProbePSQLEndpoint(ctx context.Context, config *Config) error {
	timeoutDuration, intervalDuration, err := parseDurations(config.Timeout, config.Interval)
	if err != nil {
		return err
	}
	sslMode := config.PsqlSSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	dsn := config.PsqlDsn
	if dsn == "" {
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%v connect_timeout=%d",
			config.PsqlHost, config.PsqlPort, config.PsqlUser, config.PsqlPassword, config.PsqlDatabase, sslMode, timeoutDuration)
	}
	for {
		Log.Info(fmt.Sprintf("Probing PostgreSQL endpoint host=%s port=%d timeout=%s interval=%s",
			config.PsqlHost, config.PsqlPort, config.Timeout, config.Interval))
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			Log.Debug(err.Error())
		}
		defer func() {
			if db != nil {
				db.Close()
			}
		}()
		if err != nil {
			continue
		}
		err = db.Ping()
		if err != nil {
			Log.Debug(err.Error())
		}

		if err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("cancled while waiting for %s: %w", config.PsqlHost, ctx.Err())
		default:
			time.Sleep(intervalDuration)
		}
	}
}

func ProbeMongoEndpoint(probeCtx context.Context, config *Config) error {
	timeoutDuration, intervalDuration, err := parseDurations(config.Timeout, config.Interval)
	if err != nil {
		return err
	}
	uri := mongoUriBuilder(config.MongoHost, config.MongoPort, config.MongoUser, config.MongoPassword, config.MongoDatabase, config.MongoAuthSource)
	for {
		Log.Info(fmt.Sprintf("Probing Mongo endpoint host=%s port=%d timeout=%s interval=%s",
			config.MongoHost, config.MongoPort, config.Timeout, config.Interval))

		ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
		defer cancel()

		client, err := mongo.Connect(options.Client().ApplyURI(uri))
		if err != nil {
			Log.Debug(err.Error())
		}
		defer func() { _ = client.Disconnect(context.Background()) }()

		err = client.Ping(ctx, readpref.Primary())

		if err != nil {
			Log.Debug(err.Error())
		}
		if err == nil {
			return nil
		}

		select {
		case <-probeCtx.Done():
			return fmt.Errorf("cancled while waiting for %s: %w", config.MongoHost, ctx.Err())
		default:
			time.Sleep(intervalDuration)
		}
	}
}

func parseDurations(timeout string, interval string) (time.Duration, time.Duration, error) {
	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid timeout: %w", err)
	}
	intervalDuration, err := time.ParseDuration(interval)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid interval: %w", err)
	}
	return timeoutDuration, intervalDuration, nil
}

func validateConfig(config *Config) error {
	oneOfStr := []string{config.URL, config.PsqlHost, config.MongoHost}
	oneOfSlice := [][]string{config.URLS}
	count := 0
	for _, value := range oneOfStr {
		if value != "" {
			count++
		}
	}
	for _, slice := range oneOfSlice {
		if slice != nil && len(slice) > 0 {
			count++
		}
	}
	if count == 0 {
		return fmt.Errorf("at least one of --url, --psql-host, --mongo-host, --urls must be provided")
	}
	if IsPostgresProbeActive(config) {
		return validatePostgresProbeConfig(config)
	}

	if IsHTTPProbeActive(config) {
		return validateHTTPProbeConfig(config)
	}

	if IsMongoProbeActive(config) {
		return validateMongoProbeConfig(config)
	}

	return nil
}

func IsMongoProbeActive(config *Config) bool {
	return config.MongoHost != "" || config.MongoUser != "" || config.MongoPassword != "" || config.MongoDatabase != "" || config.MongoAuthSource != ""
}

func validateMongoProbeConfig(config *Config) error {
	if config.MongoHost == "" || config.MongoPort == 0 {
		return fmt.Errorf("at least --mongo-host and --mongo-port must be provided for a mongo probe")
	}
	return nil
}

func IsHTTPProbeActive(config *Config) bool {
	return config.URL != "" || len(config.URLS) > 0
}

func validateHTTPProbeConfig(config *Config) error {
	var urls []string
	if config.URL != "" {
		urls = []string{config.URL}
	}
	if config.URLS != nil && len(config.URLS) > 0 {
		urls = append(urls, config.URLS...)
	}

	for _, _url := range urls {
		_, err := url.ParseRequestURI(_url)
		if err != nil {
			return err
		}
	}
	return nil
}

func IsPostgresProbeActive(config *Config) bool {
	return config.PsqlHost != "" || config.PsqlUser != "" || config.PsqlPassword != "" || config.PsqlDatabase != ""
}

func validatePostgresProbeConfig(config *Config) error {
	if config.PsqlHost == "" {
		return fmt.Errorf("--psql-host is required for a postgres probe")
	}
	if config.PsqlPort == 0 {
		return fmt.Errorf("--psql-port is required for a postgres probe")
	}
	if config.PsqlUser == "" {
		return fmt.Errorf("--psql-user is required for a postgres probe")
	}
	if config.PsqlPassword == "" {
		return fmt.Errorf("--psql-password is required for a postgres probe")
	}
	if config.PsqlDatabase == "" {
		return fmt.Errorf("--psql-database is required for a postgres probe")
	}
	return nil
}

func mongoUriBuilder(host string, port int, user string, password string, database string, authSource string) string {
	if port == 0 {
		port = 27017
	}
	var mongoUri strings.Builder
	mongoUri.WriteString("mongodb://")
	if user != "" && password != "" {
		mongoUri.WriteString(fmt.Sprintf("%s:%s@", url.QueryEscape(user), url.QueryEscape(password)))
	}
	mongoUri.WriteString(fmt.Sprintf("%s:%d", host, port))
	if database != "" {
		mongoUri.WriteString(fmt.Sprintf("/%s", url.QueryEscape(database)))
	}

	if authSource != "" {
		mongoUri.WriteString(fmt.Sprintf("?authSource=%s", url.QueryEscape(authSource)))
	}

	return mongoUri.String()
}

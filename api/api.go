package api

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/JacobNewton007/busha-test/internals/data"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	version   = "1.0.0"
	buildTime string
)

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	// redis_url string
}

type application struct {
	config config
	models data.Models
	logger zap.SugaredLogger
	client redis.Client
}

var (
	ErrNil = errors.New("no matching record found")
	Ctx    = context.TODO()
)

func RunApi() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server ports")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")
	flag.StringVar(&cfg.db.dsn, "redis-url", "", "Redis URL")

	// Read the connection pool settings from command-line flags into the config struct.
	// Notice the default values that we're using?
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	displayVersion := flag.Bool("version", false, "Display version and exist")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		fmt.Printf("Build time:\t%s\n", buildTime)
		os.Exit(0)
	}
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.EncoderConfig.TimeKey = "timestamp"
	loggerConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	logger, err := loggerConfig.Build()
	if err != nil {
		log.Fatal(err)
	}

	sugar := logger.Sugar()

	db, err := openDB(cfg)
	if err != nil {
		sugar.Fatal(err)
	}

	client, err := InitialRedis(cfg)
	if err != nil {
		sugar.Error(err)
	}

	defer db.Close()
	sugar.Infow("database connection pool established", "tag", "database-connection")

	app := &application{
		config: cfg,
		models: data.CommentFactory(db),
		logger: *sugar,
		client: *client,
	}

	err = app.server()
	if err != nil {
		sugar.Fatal(err)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	// Set the maximum idle timeout.
	db.SetConnMaxIdleTime(duration)

	// Create a context with a 5-second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func InitialRedis(cfg config) (*redis.Client, error) {
	if os.Getenv("APP_ENV") != "PRODUCTION" {
		err := godotenv.Load(".envrc")
		if err != nil {
			log.Println("Error loading .env file")
		}

		client := redis.NewClient(&redis.Options{
			Addr:     "127.0.0.1:6379",
			Password: "",
			DB:       0,
		})
		err = client.Ping(Ctx).Err()
		if err != nil {
			log.Fatalf("Failed to connect to redis: %s", err.Error())
		}
		return client, nil
	} else {
		redis_url := os.Getenv("REDIS_URL")
		redisUrl, _ := url.Parse(redis_url)
		redisPassword, _ := redisUrl.User.Password()
		redisOptions := redis.Options{
			Addr:     redisUrl.Host,
			Password: redisPassword,
			DB:       0,
		}

		client := redis.NewClient(&redisOptions)
		err := client.Ping(Ctx).Err()
		if err != nil {
			log.Fatalf("Failed to connect to redis: %s", err.Error())
		}
		return client, nil
	}
}

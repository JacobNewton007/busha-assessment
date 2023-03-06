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
	"strconv"
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

var (

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

type db_config struct {
	dsn          string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
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
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file")
	}
	port, _ := strconv.ParseInt(os.Getenv("PORT"), 10, 64)
	max_conns, _ := strconv.ParseInt(os.Getenv("MAX_CONNS"), 10, 64)
	idle_conns, _ := strconv.ParseInt(os.Getenv("IDLE_CONNS"), 10, 64)

	db_config := db_config{
		dsn: os.Getenv("BUSHA_DB"),
		maxOpenConns: int(max_conns),
		maxIdleConns: int(idle_conns),
		maxIdleTime:	os.Getenv("IDLE_TIME"),
	}

	cfg := config{
		port: int(port),
		env: os.Getenv("APP_ENV"),
		db: db_config,
	}

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
	fmt.Println(cfg.db.dsn)
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
		client := redis.NewClient(&redis.Options{
			Addr:     "127.0.0.1:6379",
			Password: "",
			DB:       0,
		})
		err := client.Ping(Ctx).Err()
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

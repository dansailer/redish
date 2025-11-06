package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	version     = "dev"
	showHelp    = flag.Bool("help", false, "Output this help and exit.")
	showVersion = flag.Bool("version", false, "Output version and exit.")
	uri         = flag.String("uri", "localhost:6379", "Redis server URI")
	useTls      = flag.Bool("tls", false, "Establish a secure TLS connection.")
	insecure    = flag.Bool("insecure", false, "Allow insecure TLS connection by skipping cert validation.")
	logLevel    = flag.String("logLevel", "warn", "Log level (debug, info, warn, error, fatal, panic)")
	user        = flag.String("user", "", "Username to use when connecting. Supported since Redis 6.")
	password    = flag.String("password", "", "Password to use when connecting or empty and use the REDIS_PASSWORD environment variable")
	db          = flag.Int("db", 0, "Redis database to access")
	commands    = flag.String("commands", "", "Redis commands to execute")
	tlsConfig   *tls.Config
	ctx         = context.Background()
)

func main() {
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.WarnLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if lvl, err := zerolog.ParseLevel(*logLevel); err == nil {
		zerolog.SetGlobalLevel(lvl)
	} else {
		log.Warn().Msg("Invalid log level, defaulting to 'warn'")
	}
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()

	if *showHelp {
		fmt.Println("Usage of Redish - the redis-cli:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *showVersion {
		fmt.Println("Redish - the redis-cli - version:", version)
		os.Exit(0)
	}

	if *useTls {
		tlsConfig = &tls.Config{}
		if *insecure {
			tlsConfig.InsecureSkipVerify = true
		}
	}

	if *password == "" {
		*password = os.Getenv("REDIS_PASSWORD")
	}

	client := redis.NewClient(&redis.Options{
		Addr:      *uri,
		Username:  *user,
		Password:  *password,
		DB:        *db,
		TLSConfig: tlsConfig,

		// https://github.com/redis/go-redis/issues/3536
		// Explicitly disable maintenance notifications
		// This prevents the client from sending CLIENT MAINT_NOTIFICATIONS ON
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		},
	})

	log.Info().Msg("Attempting to connect to Redis...")
	if err := client.Ping(ctx).Err(); err != nil {
		log.Error().Err(err).Msg("Failed to connect to Redis")
		os.Exit(1)
	} else {
		log.Info().Msg("Successfully connected to Redis")
	}

	if *commands != "" {
		commandArray := strings.Split(*commands, ";")
		for _, command := range commandArray {
			handleCommand(command, client)
		}
	} else {
		rl, err := readline.New("> ")
		if err != nil {
			log.Fatal().Err(err).Msg("error creating readline")
			os.Exit(1)
		}
		defer rl.Close()

		for {
			line, err := rl.Readline()
			if err != nil {
				log.Fatal().Err(err).Msg("error reading input")
				break
			}
			handleCommand(line, client)
		}
	}
}

func handleCommand(line string, client *redis.Client) {
	line = strings.TrimSpace(line)
	if line == "exit" {
		log.Info().Msg("exiting...")
		os.Exit(0)
	}

	args := strings.Fields(line)
	if len(args) == 0 {
		return
	}

	cmdArgs := make([]interface{}, len(args))
	for i, arg := range args {
		cmdArgs[i] = arg
	}

	cmd := client.Do(ctx, cmdArgs...)
	result, err := cmd.Result()
	if err != nil {
		log.Error().Err(err).Msg("Error executing command")
		return
	}
	switch v := result.(type) {
	case []interface{}:
		for _, j := range v {
			fmt.Printf("%s\n", toValueString(j))
		}
	default:
		fmt.Printf("%s\n", toValueString(result))
	}
}

func toValueString(value interface{}) string {
	switch v := value.(type) {
	case redis.Error:
		log.Error().Msg("Error executing command: " + v.Error())
	case int64:
		return fmt.Sprintf("%d", v)
	case string:
		return v
	case []byte:
		return string(v)
	case nil:
		return "nil"
	}
	return ""
}

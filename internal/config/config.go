package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

const (
	defaultRwTimeout       = 5
	defaultParsingTimeout  = 10
	defaultParsingInterval = 10
	defaultParsingChanBuff = 2
)

var (
	ErrNoDbUrl         = errors.New("missing DB_URL")
	ErrNoTelegramToken = errors.New("missing BOT_TOKEN")
	ErrNoNetAddr       = errors.New("missing ADDR")

	ErrConfigNotFound = errors.New("config file not found")
)

type Config struct {
	Net struct {
		// Address to listen to with http.
		// e.g. localhost:5000.
		Addr string

		// Read Write http timeout.
		// Represented in seconds.
		RWTimeout time.Duration
	}

	Telegram struct {
		// Telegram API bot token.
		Token string
	}

	Parsing struct {
		// Interval between parsing.
		// Represented in seconds.
		Interval time.Duration

		// Maximum amount of time for one(1) parsing.
		// Represented in seconds.
		Timeout time.Duration

		// Maximum amount of results queued
		// within output chan inside parser.
		// Could increase perfomance on high-load
		// Try to keep as small as possible.
		ChanBuff int32
	}

	Database struct {
		// Database connection string
		Url string
	}
}

func Load(path string) (*Config, error) {

	// Set full path /path/to/[name].yml
	viper.SetConfigFile(path)

	// Check if exists
	_, err := os.Stat(path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil, ErrConfigNotFound
	}

	err = viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("viper.ReadInConfig: %w", err)
	}

	dbUrl, ok := os.LookupEnv("DB_URL")
	if !ok {
		return nil, ErrNoDbUrl
	}

	netAddr, ok := os.LookupEnv("ADDR")
	if !ok {
		return nil, ErrNoNetAddr
	}

	token, ok := os.LookupEnv("BOT_TOKEN")
	if !ok {
		return nil, ErrNoTelegramToken
	}

	var netRwTimeout = viper.GetInt64("net.rw_timeout")
	if netRwTimeout == 0 {
		netRwTimeout = defaultRwTimeout
	}

	var (
		parsingInterval = viper.GetInt64("parsing.interval")
		parsingTimeout  = viper.GetInt64("parsing.timeout")
		parsingChanBuff = viper.GetInt32("parsing.chan_buff")
	)

	if parsingInterval == 0 {
		parsingInterval = defaultParsingInterval
	}

	if parsingTimeout == 0 {
		parsingTimeout = defaultParsingTimeout
	}

	if parsingChanBuff == 0 {
		parsingChanBuff = defaultParsingChanBuff
	}

	return &Config{
		Net: struct {
			Addr      string
			RWTimeout time.Duration
		}{
			Addr:      netAddr,
			RWTimeout: time.Duration(netRwTimeout) * time.Second,
		},
		Telegram: struct{ Token string }{
			Token: token,
		},
		Parsing: struct {
			Interval time.Duration
			Timeout  time.Duration
			ChanBuff int32
		}{
			Interval: time.Duration(parsingInterval) * time.Second,
			Timeout:  time.Duration(parsingTimeout) * time.Second,
			ChanBuff: parsingChanBuff,
		},
		Database: struct{ Url string }{
			Url: dbUrl,
		},
	}, nil

}

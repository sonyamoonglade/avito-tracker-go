package app

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"parser/internal/config"
	"parser/internal/domain/repositories"
	"parser/internal/domain/services"
	"parser/internal/http"
	"parser/internal/notify"
	"parser/internal/parser"
	"parser/internal/postgres"
	"parser/internal/proxy"
	"parser/internal/telegram"
	"parser/internal/timer"
	"parser/internal/urlcache"
	"syscall"
	"time"
)

func Bootstrap() error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	configPath, debug := parseFlags()

	// Read config
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	pg, err := postgres.FromConnectionString(ctx, cfg.Database.Url)
	if err != nil {
		return fmt.Errorf("postgres: %w", err)
	}

	telegram := telegram.NewTelegram(debug)
	telegramNotifier := notify.NewTelegramNotifier(telegram)

	chromedpParser, err := parser.NewChromeParser()
	if err != nil {
		return fmt.Errorf("chromedp-parser: %w", err)
	}

	ringParser := parser.NewRingParser(&parser.RingParserOptions{
		Parser:         chromedpParser,
		ParsingTimeout: cfg.Parsing.Timeout * time.Second,
		Timer:          new(timer.AppTimer),
		OutChanBuff:    cfg.Parsing.ChanBuff,
		UrlCache:       urlcache.NewUrlCache(time.Minute * 5 /* cache TTL */), // TODO: config
	})

	repositories := repositories.NewRepositories(pg)
	services := services.NewServices(repositories, telegramNotifier, ringParser)

	// Adds all URLs for parsing to ringParser
	fetcher := services.SubscriptionService.GetURLFetcher()
	if err := addInitialUrls(ctx, ringParser, fetcher); err != nil {
		return fmt.Errorf("add initial urls: %w", err)
	}

	ringParser.Run(cfg.Parsing.Interval)

	updateHandler := services.SubscriptionService.GetUpdateHandler()
	proxy := proxy.NewProxy(ringParser.Out(), updateHandler, func(err error) /* err handl. callback */ {
		// Placeholder
		// TODO: replace with proper error handler
		fmt.Printf("proxy error: %v\n", err)
	})
	// Start reading from ringParser output and executing updateHandler
	go proxy.Run()

	server := http.NewHTTPServer(&http.ServerConfig{
		Router:       http.NewMuxRouter(),
		Services:     services,
		Addr:         cfg.Net.Addr,
		WriteTimeout: cfg.Net.RWTimeout,
		ReadTimeout:  cfg.Net.RWTimeout,
	})

	go func() {
		if err := server.Run(); err != nil {
			panic(fmt.Sprintf("server unable to bootstrap: %v\n", err))
		}
	}()

	// Avoid data race
	token := cfg.Telegram.Token
	go func() {
		if err := telegram.Connect(token); err != nil {
			panic(fmt.Sprintf("unable to connect to telegram: %v\n", err))
		}
	}()

	exit := make(chan os.Signal)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)

	// Gracefull shutdown
	<-exit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		// replace with Warn
		fmt.Printf("server was unable to shutdown gracefully: %v", err)
	}

	ringParser.Close()
	pg.Close()
	telegram.Close()

	return nil
}

func addInitialUrls(ctx context.Context, ringParser *parser.RingParser, fetcher func(ctx context.Context) ([]string, error)) error {
	urls, err := fetcher(ctx)
	if err != nil {
		return err
	}

	for _, url := range urls {
		fmt.Printf("adding url: %s\n", url)
		ringParser.AddTarget(url)
	}

	return nil
}

func parseFlags() (string, bool) {
	configPath := flag.String("config", "", "path to config.yaml")
	debug := flag.Bool("debug", false, "set debug mode (more logging)")

	flag.Parse()

	if *configPath == "" {
		panic("missing config path. Use --config")
	}

	return *configPath, *debug
}

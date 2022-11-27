package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"parser/internal/domain/repositories"
	"parser/internal/domain/services"
	"parser/internal/http"
	"parser/internal/notify"
	"parser/internal/parser"
	"parser/internal/postgres"
	"parser/internal/telegram"
	"parser/internal/timer"
	"syscall"
	"time"
)

func Bootstrap() error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// TODO: get from config
	connString := os.Getenv("DB_URL")
	// move to config
	if connString == "" {
		return errors.New("DB_URL missing")
	}
	pg, err := postgres.FromConnectionString(ctx, connString)
	if err != nil {
		return err
	}
	fmt.Println("database has connected")

	telegram := telegram.NewTelegram()
	telegramNotifier := notify.NewTelegramNotifier(telegram)

	chromedpParser, err := parser.NewChromeParser()
	if err != nil {
		return err
	}

	// TODO: move to config
	waitingqLen := 10
	parsingTimeout := time.Second * 10
	ringParser := parser.NewRingParser(chromedpParser, new(timer.AppTimer), parsingTimeout, waitingqLen)

	repositories := repositories.NewRepositories(pg)
	services := services.NewServices(repositories, telegramNotifier, ringParser)

	// Adds all URLs for parsing to ringParser
	fetcher := services.SubscriptionService.GetURLFetcher()
	if err := addInitialUrls(ctx, ringParser, fetcher); err != nil {
		return err
	}

	// TODO: get timeout from config
	parsingInterval := time.Second * 20
	ringParser.Run(parsingInterval)

	updateHandler := services.SubscriptionService.GetUpdateHandler()
	proxy := parser.NewProxy(ringParser.Out(), updateHandler, func(err error) {
		// Placeholder
		fmt.Printf("proxy error: %v\n", err)
	})
	go proxy.Run()

	// TODO: introduce defaults for timeout and config..
	server := http.NewHTTPServer(&http.ServerConfig{
		Router:       http.NewMuxRouter(),
		Addr:         "127.0.0.1:8000",
		Services:     services,
		WriteTimeout: time.Second * 10,
		ReadTimeout:  time.Second * 10,
	})

	go func() {
		if err := server.Run(); err != nil {
			panic(fmt.Sprintf("server unable to bootstrap: %v\n", err))
		}
	}()
	fmt.Println("http server is up")

	// TODO: get from config
	token := os.Getenv("BOT_TOKEN")
	go func() {
		if err := telegram.Connect(token); err != nil {
			panic(fmt.Sprintf("unable to connect to telegram: %v\n", err))
		}
	}()
	fmt.Println("telegram is connected")

	exit := make(chan os.Signal)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGKILL)

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

	fmt.Println("shutdown gracefully")

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

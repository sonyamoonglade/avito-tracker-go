package main

import (
	"context"
	"fmt"
	"log"
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

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// TODO: get from config
	connString := os.Getenv("DB_URL")
	// move to config
	if connString == "" {
		log.Fatalf("DB_URL is not provied")
	}
	pg, err := postgres.FromConnectionString(ctx, connString)
	if err != nil {
		log.Fatalf("err: %s", err.Error())
	}
	fmt.Println("database has connected")

	telegram := telegram.NewTelegram()
	telegramNotifier := notify.NewTelegramNotifier(telegram)

	chromedpParser, err := parser.NewChromeParser()
	if err != nil {
		log.Fatal(err)
	}

	waitingqLen := 10
	ringParser := parser.NewRingParser(chromedpParser, new(timer.AppTimer), waitingqLen)

	repositories := repositories.NewRepositories(pg)
	services := services.NewServices(repositories, telegramNotifier, ringParser)

	// Adds all URLs for parsing to ringParser
	fetcher := services.SubscriptionService.GetAllURLs
	if err := addInitialUrls(ctx, ringParser, fetcher); err != nil {
		log.Fatal(err)
	}

	// TODO: get timeout from config
	go ringParser.Run(time.Second * 10)

	proxy := parser.NewProxy(ringParser.Out(), services.SubscriptionService.HandleParsingResult, func(err error) {
		// Placeholder

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
			log.Fatal(err.Error())
		}

	}()

	fmt.Println("http server is up")

	// TODO: get from config
	token := os.Getenv("BOT_TOKEN")
	go func() {
		if err := telegram.Connect(token); err != nil {
			log.Fatalf("error connecting to telegram: %s", err.Error())
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
		log.Printf("server was unable to shutdown gracefully: %v", err)
	}

	ringParser.Close()
	pg.Close()
	telegram.Close()

	fmt.Println("shutdown gracefully")
}

// TODO: think of appropriate place...
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

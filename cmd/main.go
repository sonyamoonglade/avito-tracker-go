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

	chromedpParser, err := parser.NewChromeParser()
	if err != nil {
		log.Fatal(err)
	}

	urls := []string{
		"https://www.avito.ru/izhevsk/telefony/iphone_6_i_6s_2607330465",
		"https://www.avito.ru/izhevsk/telefony/prodaetsya_afon_6s_na_zapchasti_2674592866",
		"https://www.avito.ru/izhevsk/telefony/telefon_iphone_6_2690077250",
	}

	ringParser := parser.NewRingParser(chromedpParser, new(timer.AppTimer), urls...)

	// TODO: get timeout from config
	go ringParser.Run(time.Second * 10)

	// data provider: ringParser.Out()

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

	repositories := repositories.NewRepositories(pg)
	services := services.NewServices(repositories, telegramNotifier)

	// TODO: introduce defaults for timeout and config..
	server := http.NewHTTPServer(&http.ServerConfig{
		Router:       http.NewMuxRouter(),
		Addr:         "127.0.0.1:8000",
		Services:     services,
		WriteTimeout: time.Second * 10,
		ReadTimeout:  time.Second * 10,
	})

	go func() {
		err := server.Run()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	fmt.Println("http server is up")

	// TODO: get from config
	token := os.Getenv("BOT_TOKEN")
	go func() {
		err := telegram.Connect(token)
		if err != nil {
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

	ringParser.Stop()
	pg.Close()
	telegram.Close()

	fmt.Println("shutdown gracefully")
}

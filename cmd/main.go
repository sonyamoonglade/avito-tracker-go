package main

import (
	"context"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"parser/internal/http"
	"parser/internal/notify"
	"parser/internal/parser"
	"parser/internal/telegram"
	"parser/internal/timer"
	"parser/pkg/postgres"
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
	_ = telegramNotifier

	// TODO: get from config
	token := os.Getenv("BOT_TOKEN")
	go func() {
		err := telegram.Connect(token)
		if err != nil {
			log.Fatalf("error connecting to telegram: %s", err.Error())
		}
	}()
	fmt.Println("telegram is connected")

	// TODO: introduce defaults for timeout and config..
	server := http.NewHTTPServer(http.NewMuxRouter(), "localhost:8000", time.Second*10, time.Second*10)
	go func() {
		err := server.Run()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	fmt.Println("http server is up")
	exit := make(chan os.Signal)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGKILL)

	// gracefull shutdown
	<-exit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	pg.Close()

	if err := server.Shutdown(shutdownCtx); err != nil {
		// replace with Warn
		log.Printf("server was unable to shutdown gracefully: %v", err)
	}

	fmt.Println("shutdown gracefully")
}

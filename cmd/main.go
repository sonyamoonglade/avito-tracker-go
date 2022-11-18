package main

import (
	"context"
	"log"
	"os"
	"parser/internal/domain/advert"
	"parser/internal/domain/subscriber"
	"parser/internal/domain/subscription"
	"parser/internal/http"
	"parser/internal/http/routing"
	"parser/internal/parser"
	"parser/pkg/postgres"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	url := "https://www.avito.ru/izhevsk/telefony/prodaetsya_afon_6s_na_zapchasti_2674592866"
	_ = url

	chromedpParser, err := parser.NewChromeParser(time.Second * 15)
	if err != nil {
		log.Fatal(err)
	}

	connString := os.Getenv("DB_URL")
	pg, err := postgres.FromConnectionString(ctx, connString)
	if err != nil {
		log.Fatalf("err: %s", err.Error())
	}

	// services
	advertService := advert.NewService(pg)
	subscriberService := subscriber.NewService(pg)
	subscriptionService := subscription.NewService(pg)

	_ = advertService
	_ = subscriberService
	_ = chromedpParser

	// handler
	subscriptionHandler := subscription.NewHandler(subscriptionService)

	httpHandlers := http.NewHttpHandlers(subscriptionHandler)

	router := routing.New()
	router.InitializeHttpEndpoints(httpHandlers)

	server := http.NewServer(router.Handler())

	err = server.Run("localhost:8000", time.Second*10, time.Second*10)
	if err != nil {
		log.Fatal(err.Error())
	}
}

package main

import (
	"log"
	"parser/internal/app"
)

func main() {
	if err := app.Bootstrap(); err != nil {
		log.Fatal(err.Error())
	}
}

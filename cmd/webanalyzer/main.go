package main

import (
	"log"
	"net/http"
	"webanalyzer/internal/api/v1/router"
)

func main() {
	r := router.New()

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	log.Println("Server started..")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Antony8720/url-shortener/internal/app"
	"github.com/Antony8720/url-shortener/internal/config"
	"github.com/Antony8720/url-shortener/internal/storage"
)

func main() {
	log.Print("url-shortener: Enter main()")
	cfg := config.New()
	storage, err := storage.New(cfg)
	if err != nil {
		fmt.Println(err)
		return
	}
	baseURL := cfg.BaseURL
	DBAddress := cfg.DBAddress
	log.Fatal(http.ListenAndServe(cfg.Address, app.MainRouter(storage, baseURL, DBAddress)))
}

package config

import (
	"flag"
	"os"
)

type Cfg struct {
	Filepath  string
	Address   string
	BaseURL   string
	DBAddress string
}

func New() Cfg {
	cfg := Cfg{}
	flag.StringVar(&cfg.Filepath, "f", "", "path to the file fith shortened URLs")
	flag.StringVar(&cfg.Address, "a", "", "start address of the HTTP server")
	flag.StringVar(&cfg.BaseURL, "b", "", "base address of the resulting shortened URL")
	flag.StringVar(&cfg.DBAddress, "d", "", "DB connection address")
	flag.Parse()
	cfg.chooseFilepath()
	cfg.chooseAddress()
	cfg.chooseBaseURL()
	cfg.chooseDBAddress()
	return cfg
}

func (cfg *Cfg) chooseFilepath() {
	if cfg.Filepath != "" {
		return
	}
	filepath, ok := os.LookupEnv("FILE_STORAGE_PATH")
	if !ok {
		return
	}
	cfg.Filepath = filepath
}

func (cfg *Cfg) chooseAddress() {
	if cfg.Address != "" {
		return
	}
	address, ok := os.LookupEnv("SERVER_ADDRESS")
	if !ok {
		address = ":8080"
	}
	cfg.Address = address
}

func (cfg *Cfg) chooseBaseURL() {
	if cfg.BaseURL != "" {
		return
	}
	bu, ok := os.LookupEnv("BASE_URL")
	if !ok {
		bu = "http://localhost:8080"
	}
	cfg.BaseURL = bu
}

func (cfg *Cfg) chooseDBAddress() {
	if cfg.DBAddress != "" {
		return
	}
	dba := os.Getenv("DATABASE_DSN")
	cfg.DBAddress = dba
}

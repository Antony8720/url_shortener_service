package storage

import (
	"github.com/Antony8720/url-shortener/internal/config"
)

func New(cfg config.Cfg) (URLStorage, error) {
	if len(cfg.DBAddress) > 0 {
		return NewDatabaseStorage(cfg.DBAddress)
	}
	if len(cfg.Filepath) == 0 {
		return NewDataStorage(), nil
	}
	storage, err := NewFileStorage(cfg.Filepath)
	if err != nil {
		return nil, err
	}
	err = storage.LoadingDataFromFile()
	if err != nil {
		return nil, err
	}
	return storage, nil
}

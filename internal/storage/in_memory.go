package storage

import (
	"github.com/google/uuid"
	"sync"
)

type URLStorage interface {
	Get(string) (string, bool)
	Set(uuid.UUID, string, string) error
	GetHistory(uuid.UUID) (map[string]string, error)
}

type DataStorage struct {
	sync.RWMutex
	cache   map[string]string
	history map[uuid.UUID]map[string]string
}

func NewDataStorage() *DataStorage {
	return &DataStorage{
		cache:   make(map[string]string),
		history: make(map[uuid.UUID]map[string]string),
	}
}

func (ds *DataStorage) Get(key string) (string, bool) {
	ds.RLock()
	defer ds.RUnlock()
	value, ok := ds.cache[key]
	return value, ok
}

func (ds *DataStorage) Set(userID uuid.UUID, key, value string) error {
	ds.Lock()
	defer ds.Unlock()
	if userID != uuid.Nil {
		if _, ok := ds.history[userID]; !ok {
			ds.history[userID] = map[string]string{}
		}
		ds.history[userID][key] = value
	}
	ds.cache[key] = value
	return nil
}

func (ds *DataStorage) GetHistory(uuid uuid.UUID) (map[string]string, error) {
	result := ds.history[uuid]
	return result, nil
}

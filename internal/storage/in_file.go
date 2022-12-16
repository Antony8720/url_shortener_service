package storage

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/google/uuid"
)

type FileStorage struct {
	file    *os.File
	storage DataStorage
}

type url struct {
	UserID uuid.UUID `json:"userID,omitempty"`
	Short  string    `json:"short"`
	Long   string    `json:"long"`
}

func NewFileStorage(filename string) (*FileStorage, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return &FileStorage{}, err
	}
	return &FileStorage{
		file:    file,
		storage: *NewDataStorage(),
	}, nil
}

func (f *FileStorage) Close() error {
	return f.file.Close()
}

func (f *FileStorage) LoadingDataFromFile() error {
	scanner := bufio.NewScanner(f.file)
	var url url
	for scanner.Scan() {
		data := scanner.Bytes()
		err := json.Unmarshal(data, &url)
		if err != nil {
			return err
		}
		f.storage.cache[url.Short] = url.Long
		if url.UserID != uuid.Nil {
			if _, ok := f.storage.history[url.UserID]; !ok {
				f.storage.history[url.UserID] = map[string]string{}
			}
			f.storage.history[url.UserID][url.Short] = url.Long
		}
	}
	return nil
}

func (f *FileStorage) Get(short string) (long string, ok bool) {
	return f.storage.Get(short)
}

func (f *FileStorage) Set(userID uuid.UUID, short, long string) error {
	err := f.storage.Set(userID, short, long)
	if err != nil {
		return err
	}
	err = f.WriteURLInFile(userID, short, long)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileStorage) WriteURLInFile(userID uuid.UUID, short, long string) error {
	s := url{
		UserID: userID,
		Short:  short,
		Long:   long,
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = f.file.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileStorage) GetHistory(userID uuid.UUID) (map[string]string, error) {
	return f.storage.history[userID], nil
}

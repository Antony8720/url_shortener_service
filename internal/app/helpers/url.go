package helpers

import (
	"errors"

	"github.com/Antony8720/url-shortener/internal/app/violationerror"
	"github.com/Antony8720/url-shortener/internal/storage"
	"github.com/Antony8720/url-shortener/internal/utils"
	"github.com/google/uuid"
)

func EncodeURL(userID uuid.UUID, baseURL string, urlStorage storage.URLStorage) (string, error){
	encURL := utils.RandURL()
	for{
		_, ok := urlStorage.Get(encURL)
		if ok{
			encURL = utils.RandURL()
		} else {
			break
		}
	}
	err := urlStorage.Set(userID, encURL, baseURL)
	if err != nil{
		var uve *violationerror.UniqueViolationError
		if errors.As(err, &uve){
			if err, ok := err.(*violationerror.UniqueViolationError); ok{
				return err.Short, err
			}
		}
		return "", err
	}
	return encURL, err
}

func DecodeURL(encURL string, urlStorage storage.URLStorage) (baseURL string, ok bool){
	return urlStorage.Get(encURL)
}
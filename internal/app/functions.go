package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
	"github.com/Antony8720/url-shortener/internal/app/violationerror"
	"github.com/Antony8720/url-shortener/internal/app/helpers"
	"github.com/Antony8720/url-shortener/internal/storage"
	"github.com/Antony8720/url-shortener/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

type RequestJSON struct {
	URL string `json:"url"`
}

type ResponseJSON struct {
	Result string `json:"result"`
}

type result struct {
	Short string `json:"short_url"`
	Long  string `json:"original_url"`
}

type InputBatch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type OutputBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func SaveLongURL(storage storage.URLStorage, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "400 bad request", http.StatusBadRequest)
			return
		}

		defer r.Body.Close()
		longURL := string(b)
		var writeBody = func(b []byte) {
			w.Header().Set("content-type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusCreated)
			w.Write(b)
		}
		u, ok := GetRequestUser(r)
		if !ok {
			u = user.User{UserID: uuid.Nil}
		}

		encURL, err := helpers.EncodeURL(u.UserID, longURL, storage)
		if err != nil {
			var uve *violationerror.UniqueViolationError

			if !errors.As(err, &uve) {
				http.Error(w, "400 page not found", http.StatusBadRequest)
				return
			}

			writeBody = func(b []byte) {
				w.Header().Set("content-type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusConflict)
				w.Write(b)
			}

		}

		var fullEncURL string
		if baseURL == "" {
			fullEncURL = fmt.Sprintf("http://%s/%s", r.Host, encURL)
		} else {
			fullEncURL = fmt.Sprintf("%s/%s", baseURL, encURL)
		}

		writeBody([]byte(fullEncURL))
	}
}

func SaveJSONLongURL(storage storage.URLStorage, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := RequestJSON{}
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "400 page not found", http.StatusBadRequest)
			return
		}

		defer r.Body.Close()
		if err := json.Unmarshal(b, &req); err != nil {
			http.Error(w, "400 page not found", http.StatusBadRequest)
			return
		}

		u, ok := GetRequestUser(r)
		if !ok {
			u = user.User{UserID: uuid.Nil}
		}

		var writeBody = func(b []byte) {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write(b)
		}

		encURL, err := helpers.EncodeURL(u.UserID, req.URL, storage)
		if err != nil {
			var uve *violationerror.UniqueViolationError

			if !errors.As(err, &uve) {
				http.Error(w, "400 page not found", http.StatusBadRequest)
				return
			}

			writeBody = func(b []byte) {
				w.Header().Set("content-type", "application/json")
				w.WriteHeader(http.StatusConflict)
				w.Write(b)
			}

		}

		var fullEncURL string
		if baseURL == "" {
			fullEncURL = fmt.Sprintf("http://%s/%s", r.Host, encURL)
		} else {
			fullEncURL = fmt.Sprintf("%s/%s", baseURL, encURL)
		}

		resp := ResponseJSON{Result: fullEncURL}
		respBody, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "400 page not found", http.StatusBadRequest)
			return
		}

		writeBody(respBody)
	}
}

func RedirectToOriginalURL(storage storage.URLStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlPart := chi.URLParam(r, "url")
		originalURL, ok := helpers.DecodeURL(urlPart, storage)
		if !ok {
			http.Error(w, "400 page not found", http.StatusBadRequest)
			return
		}

		w.Header().Set("content-type", "text/plain; charset=utf-8")
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func GetRequestUser(r *http.Request) (u user.User, ok bool) {
	ck, err := r.Cookie("Authorization")
	if err != nil {
		return user.User{}, false
	}

	err = u.UserDecryptDecodeFromString(ck.Value)
	if err != nil {
		return user.User{}, false
	}

	return u, true
}

func GetUserURLs(storage storage.URLStorage, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := GetRequestUser(r)
		if !ok {
			u = user.User{UserID: uuid.Nil}
		}

		all, err := storage.GetHistory(u.UserID)
		if err != nil {
			http.Error(w, "400 page not found", http.StatusBadRequest)
			return
		}

		if len(all) == 0 {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusNoContent)

			return
		}
		var res []result
		for short, long := range all {
			res = append(res, result{
				Short: fmt.Sprintf("%s/%s", baseURL, short),
				Long:  long,
			})
		}

		b, err := json.MarshalIndent(res, "", " ")
		if err != nil {
			http.Error(w, "400 marshalling error", http.StatusBadRequest)
		}

		b = append(b, '\n')
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}
}

func Ping(DBAddress string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pgxConfig, err := pgxpool.ParseConfig(DBAddress)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		pgxConnPool, err := pgxpool.ConnectConfig(context.TODO(), pgxConfig)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer pgxConnPool.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := pgxConnPool.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

func SaveBatch(storage storage.URLStorage, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := GetRequestUser(r)
		if !ok {
			u = user.User{UserID: uuid.Nil}
		}

		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "400 page not found", http.StatusBadRequest)
			return
		}

		defer r.Body.Close()
		var ib []InputBatch
		err = json.Unmarshal(b, &ib)
		if err != nil {
			http.Error(w, "400 page not found", http.StatusBadRequest)
			return
		}

		bo := make([]OutputBatch, 0, len(ib))
		var fullEncURL string
		for _, batch := range ib {
			encURL, err := helpers.EncodeURL(u.UserID, batch.OriginalURL, storage)
			if err != nil {
				http.Error(w, "400 page not found", http.StatusBadRequest)
				return
			}
			if baseURL == "" {
				fullEncURL = fmt.Sprintf("http://%s/%s", r.Host, encURL)
			} else {
				fullEncURL = fmt.Sprintf("%s/%s", baseURL, encURL)
			}
			bo = append(bo, OutputBatch{
				CorrelationID: batch.CorrelationID,
				ShortURL:      fullEncURL,
			})
			fmt.Print(batch)
		}

		result, err := json.MarshalIndent(bo, "", " ")
		if err != nil {
			http.Error(w, "400 page not found", http.StatusBadRequest)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(result)
	}
}

package app

import (
	"compress/gzip"
	"github.com/Antony8720/url-shortener/internal/user"
	"net/http"
	"time"
)

func checkingCompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "400 bad request", http.StatusBadRequest)
			}
			r.Body = gz
		}

		next.ServeHTTP(w, r)

	})
}

func CookieAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expires := time.Now().AddDate(0, 1, 0)
		rck, err := r.Cookie("Authorization")
		if err != nil && err != http.ErrNoCookie {
			http.Error(w, "400 cookie error", http.StatusBadRequest)
			return
		}

		if err == nil {
			rck.Expires = expires
			http.SetCookie(w, rck)
			next.ServeHTTP(w, r)
			return
		}

		ck := http.Cookie{
			Name:    "Authorization",
			Path:    "/",
			Expires: expires,
		}

		u := user.New()
		enu, err := u.UserEncryptEncodeToString()
		if err != nil {
			http.Error(w, "400 encoding error", http.StatusBadRequest)
			return
		}

		ck.Value = enu
		r.AddCookie(&ck)
		http.SetCookie(w, &ck)
		next.ServeHTTP(w, r)
	})
}

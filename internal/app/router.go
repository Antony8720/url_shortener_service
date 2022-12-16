package app

import (
	"compress/flate"
	"net/http"

	"github.com/Antony8720/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func MainRouter(storage storage.URLStorage, baseURL, DBAddress string) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.AllowContentEncoding("gzip"))
	compressor := middleware.NewCompressor(flate.DefaultCompression)
	r.Use(compressor.Handler)
	r.Use(checkingCompressionMiddleware)
	r.Use(CookieAuthorization)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("404 page not found"))
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(405)
		w.Write([]byte("405 Method not allowed"))
	})

	r.Route("/", func(r chi.Router) {
		r.Post("/", SaveLongURL(storage, baseURL))
		r.Get("/ping", Ping(DBAddress))

		r.Route("/api", func(r chi.Router) {
			r.Route("/shorten", func(r chi.Router) {
				r.Post("/", SaveJSONLongURL(storage, baseURL))
				r.Post("/batch", SaveBatch(storage, baseURL))
			})
			r.Get("/user/urls", GetUserURLs(storage, baseURL))
		})

		r.Route("/{url}", func(r chi.Router) {
			r.Get("/", RedirectToOriginalURL(storage))
		})
	})

	return r
}

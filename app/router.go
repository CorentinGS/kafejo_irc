package app

import (
	"log/slog"
	"net/http"

	"github.com/corentings/kafejo-books/app/handlers"
	"github.com/corentings/kafejo-books/assets"
	"github.com/corentings/kafejo-books/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// RegisterRoutes registers the API routes for the repertoire service.
func RegisterRoutes(s *Server) {
	pageHandler := handlers.NewPageHandler()
	bookHandler := handlers.NewBookHandler()
	chatHandler := handlers.NewChatHandler()

	r := s.Router

	if s.Config.Level == "debug" {
		r.Use(middleware.Logger) // <--<< Logger should come before Recoverer
	}

	r.Use(middleware.Heartbeat("/health"))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Origin", "Accept"},
		MaxAge:         config.CorsMaxAge, // Maximum value not ignored by any of major browsers
	}))

	r.Use(middleware.Compress(config.CompressionLevel, "text/html", "text/css"))

	r.Use(CSPMiddleware)

	r.Use(middleware.Recoverer)

	r.Mount("/debug", middleware.Profiler())

	r.Route("/static", func(r chi.Router) {
		r.Use(StaticAssetsCacheControlMiddleware)
		fs := assets.FileServer()

		r.Handle("/*", http.StripPrefix("/static/", fs))
	})

	r.Get("/", pageHandler.HandleGetIndex())

	r.Get("/books", bookHandler.HandleGetBooks())

	r.Get("/chat", pageHandler.HandleGetChat())

	r.Get("/chat/live", chatHandler.HandleGetChatLive())

	r.Post("/chat/send", chatHandler.HandlePostChatSend())

	r.Get("/robots.txt", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write(assets.RobotsTxt())
		if err != nil {
			slog.Warn("Failed to write robots.txt", slog.String("error", err.Error()))
		}
	})

	r.Get("/sitemap.xml", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, err := w.Write(assets.SitemapXML())
		if err != nil {
			slog.Warn("Failed to write sitemap.xml", slog.String("error", err.Error()))
		}
	})
}

func StaticAssetsCacheControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set Cache-Control header
		w.Header().Set("Cache-Control", "public, max-age=31536000")

		next.ServeHTTP(w, r)
	})
}

func StaticPageCacheControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set Cache-Control header
		w.Header().Set("Cache-Control", "public, max-age=0, must-revalidate")

		next.ServeHTTP(w, r)
	})
}

func CSPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set Content-Security-Policy header
		cspHeader := "default-src 'self'; connect-src 'self' *.corentings.dev ; script-src 'self' *.corentings.dev ; style-src 'self' 'unsafe-inline' cdn.jsdelivr.net cdn.simplecss.org"

		w.Header().Set("Content-Security-Policy", cspHeader)

		next.ServeHTTP(w, r)
	})
}

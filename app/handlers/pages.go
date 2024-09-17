package handlers

import (
	"log/slog"
	"net/http"

	"github.com/corentings/kafejo-books/app/views/page"
)

type PageHandler struct{}

func NewPageHandler() *PageHandler {
	return &PageHandler{}
}

func (p *PageHandler) HandleGetIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hero := page.Index()

		index := page.IndexPage("Kafejo IRC", hero)

		if err := Render(w, r, http.StatusOK, index); err != nil {
			slog.ErrorContext(r.Context(), "error rendering index page", slog.String("error", err.Error()))
		}
	}
}

func (p *PageHandler) HandleGetChat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hero := page.Chat()

		chat := page.ChatPage("Kafejo Chat", hero)

		if err := Render(w, r, http.StatusOK, chat); err != nil {
			slog.ErrorContext(r.Context(), "error rendering chat page", slog.String("error", err.Error()))
		}
	}
}

func (p *PageHandler) HandleGetLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		password := r.FormValue("password")
		if password == "" {
			http.Error(w, "Password cannot be empty", http.StatusBadRequest)
			return
		}

		if password != "password" {
			http.Error(w, "Invalid password", http.StatusUnauthorized)
			return
		}

		if err := Redirect(w, r, "/chat", http.StatusSeeOther); err != nil {
			slog.ErrorContext(r.Context(), "error redirecting to chat", slog.String("error", err.Error()))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

package handlers

import (
	"log/slog"
	"net/http"

	"github.com/corentings/kafejo-books/app/views/components"
)

type BookHandler struct{}

func NewBookHandler() *BookHandler {
	return &BookHandler{}
}

func (p *BookHandler) HandleGetBooks() http.HandlerFunc {
	books := map[string]string{
		"1": "Welcome to Kafejo Books",
		"2": "Corentin GS's Manual",
	}
	return func(w http.ResponseWriter, r *http.Request) {
		bookList := components.BooksList(books)

		if err := Render(w, r, http.StatusOK, bookList); err != nil {
			slog.ErrorContext(r.Context(), "error rendering books list", slog.String("error", err.Error()))
		}
	}
}

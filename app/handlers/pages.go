package handlers

import (
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/corentings/kafejo-books/app/views/page"
	"github.com/gorilla/securecookie"
)

var (
	globalSecureCookie *securecookie.SecureCookie
)

// InitSecureCookie initializes the global secure cookie
func InitSecureCookie(hashKey, blockKey []byte) {
	globalSecureCookie = securecookie.New(hashKey, blockKey)
}

type PageHandler struct {
	// Remove the secureCookie field
}

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
		var username string

		// Try to get the username from the secure cookie
		if cookie, err := r.Cookie("username"); err == nil {
			if err = globalSecureCookie.Decode("username", cookie.Value, &username); err != nil {
				slog.ErrorContext(r.Context(), "Failed to decode username cookie", slog.String("error", err.Error()))
			}
		}

		// If username is empty, generate a new one and set the cookie
		if username == "" {
			username = generateRandomUsername()
			encoded, err := globalSecureCookie.Encode("username", username)
			if err == nil {
				http.SetCookie(w, &http.Cookie{
					Name:     "username",
					Value:    encoded,
					Path:     "/",
					MaxAge:   86400 * 30, // 30 days
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteStrictMode,
				})
			} else {
				slog.ErrorContext(r.Context(), "Failed to encode username cookie", slog.String("error", err.Error()))
			}
		}

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

func generateRandomUsername() string {
	// Implement your random username generation logic here
	return "User" + strconv.Itoa(rand.Intn(10000))
}

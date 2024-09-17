package handlers

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/corentings/kafejo-books/app/views/page"
	"github.com/corentings/kafejo-books/domain"
)

type ChatHandler struct {
	clients     map[string]chan domain.Message
	clientsLock sync.RWMutex
	userCount   int64 // Use atomic operations for thread-safe access
}

func NewChatHandler() *ChatHandler {
	return &ChatHandler{
		clients: make(map[string]chan domain.Message),
	}
}

func (c *ChatHandler) HandleGetChatLive() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Fetch username from secure cookie
		var username string
		if cookie, err := r.Cookie("username"); err == nil {
			if err = globalSecureCookie.Decode("username", cookie.Value, &username); err != nil {
				slog.ErrorContext(r.Context(), "Failed to decode username cookie", slog.String("error", err.Error()))
				username = "Anonymous"
			}
		} else {
			username = "Anonymous"
		}

		messageChan := make(chan domain.Message)
		c.addClient(username, messageChan)
		defer c.removeClient(username)

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		// Create a ticker that ticks every 10 seconds
		userCountTicker := time.NewTicker(10 * time.Second)
		defer userCountTicker.Stop()

		for {
			select {
			case message := <-messageChan:
				// Use the username when creating the message component
				messageComponent := page.Message(message.Author, message.Content)

				// Create a buffer to render the component
				buf := &bytes.Buffer{}
				err := messageComponent.Render(r.Context(), buf)
				if err != nil {
					http.Error(w, "Rendering error", http.StatusInternalServerError)
					return
				}

				// Write the SSE formatted message
				fmt.Fprintf(w, "event: chat\ndata: %s\n\n", buf.String())
				flusher.Flush()

			case <-userCountTicker.C:
				// Get the current user count
				count := atomic.LoadInt64(&c.userCount)

				// Create the user count component
				userCountComponent := page.ConnectedUsers(count)

				// Render the component
				buf := &bytes.Buffer{}
				err := userCountComponent.Render(r.Context(), buf)
				if err != nil {
					slog.ErrorContext(r.Context(), "Error rendering user count", slog.String("error", err.Error()))
					continue
				}

				// Send the user count as an SSE event
				fmt.Fprintf(w, "event: users\ndata: %s\n\n", buf.String())
				flusher.Flush()

			case <-r.Context().Done():
				return
			}
		}
	}
}

func (c *ChatHandler) HandlePostChatSend() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		content := r.FormValue("message")
		if content == "" {
			http.Error(w, "Message cannot be empty", http.StatusBadRequest)
			return
		}

		// Fetch username from secure cookie
		var username string
		if cookie, err := r.Cookie("username"); err == nil {
			if err = globalSecureCookie.Decode("username", cookie.Value, &username); err != nil {
				slog.ErrorContext(r.Context(), "Failed to decode username cookie", slog.String("error", err.Error()))
				username = "Anonymous"
			}
		} else {
			username = "Anonymous"
		}

		message := domain.Message{
			Author:  username,
			Content: content,
		}

		c.broadcastMessage(message)

		chatInput := page.ChatInput()
		if err := chatInput.Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering chat input", http.StatusInternalServerError)
			return
		}
	}
}

func (c *ChatHandler) addClient(username string, messageChan chan domain.Message) {
	c.clientsLock.Lock()
	defer c.clientsLock.Unlock()
	if _, exists := c.clients[username]; !exists {
		c.clients[username] = messageChan
		atomic.AddInt64(&c.userCount, 1)
	}
}

func (c *ChatHandler) removeClient(username string) {
	c.clientsLock.Lock()
	defer c.clientsLock.Unlock()
	if _, exists := c.clients[username]; exists {
		close(c.clients[username])
		delete(c.clients, username)
		atomic.AddInt64(&c.userCount, -1)
	}
}

func (c *ChatHandler) broadcastMessage(message domain.Message) {
	c.clientsLock.RLock()
	defer c.clientsLock.RUnlock()
	for _, ch := range c.clients {
		select {
		case ch <- message:
		default:
			// If the client's channel is full, skip it
		}
	}
}

func (c *ChatHandler) HandleGetUserCount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		count := atomic.LoadInt64(&c.userCount)
		w.Header().Set("Content-Type", "application/json")

		connectedUsers := page.ConnectedUsers(count)
		if err := connectedUsers.Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering connected users", http.StatusInternalServerError)
			return
		}
	}
}

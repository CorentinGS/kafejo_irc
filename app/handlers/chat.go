package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/corentings/kafejo-books/app/views/page"
)

type ChatHandler struct {
	clients     map[string]chan string
	clientsLock sync.RWMutex
	userCount   int64 // Use atomic operations for thread-safe access
}

func NewChatHandler() *ChatHandler {
	return &ChatHandler{
		clients: make(map[string]chan string),
	}
}

func (c *ChatHandler) HandleGetChatLive() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		messageChan := make(chan string)
		c.addClient("", messageChan)
		defer c.removeClient("")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}
		for {
			select {
			case message := <-messageChan:
				messageComponent := page.Message(message)

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

		message := r.FormValue("message")
		if message == "" {
			http.Error(w, "Message cannot be empty", http.StatusBadRequest)
			return
		}

		c.broadcastMessage(message)

		chatInput := page.ChatInput()
		if err := chatInput.Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering chat input", http.StatusInternalServerError)
			return
		}
	}
}

func (c *ChatHandler) addClient(username string, messageChan chan string) {
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

func (c *ChatHandler) broadcastMessage(message string) {
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

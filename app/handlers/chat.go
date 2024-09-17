package handlers

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/corentings/kafejo-books/app/views/page"
)

type ChatHandler struct {
	messageChan chan string
}

func NewChatHandler() *ChatHandler {
	return &ChatHandler{
		messageChan: make(chan string),
	}
}

func (c *ChatHandler) HandleGetChatLive() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		for {
			select {
			case message := <-c.messageChan:
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

		c.SendMessage(message)

		chatInput := page.ChatInput()
		if err := chatInput.Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering chat input", http.StatusInternalServerError)
			return
		}
	}
}

func (c *ChatHandler) SendMessage(message string) {
	c.messageChan <- message
}

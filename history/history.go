package history

import (
	"sync"

	"github.com/dennislapchenko/grok-telegram-bot/models"
)

const maxHistory = 100

type ChatHistory struct {
	mu sync.Mutex
	h  map[int64][]models.ChatMessage
}

func NewChatHistory() *ChatHistory {
	return &ChatHistory{h: make(map[int64][]models.ChatMessage)}
}

func (ch *ChatHistory) Add(chatID int64, msg models.ChatMessage) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	ch.h[chatID] = append(ch.h[chatID], msg)
	if len(ch.h[chatID]) > maxHistory {
		ch.h[chatID] = ch.h[chatID][1:]
	}
}

func (ch *ChatHistory) Get(chatID int64) []models.ChatMessage {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	return append([]models.ChatMessage{}, ch.h[chatID]...)
}

func (ch *ChatHistory) Clear(chatID int64) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	delete(ch.h, chatID)
}

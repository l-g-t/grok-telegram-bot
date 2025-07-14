package bot

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unicode/utf8"

	"github.com/dennislapchenko/grok-telegram-bot/config"
	"github.com/dennislapchenko/grok-telegram-bot/grok"
	"github.com/dennislapchenko/grok-telegram-bot/history"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

const (
	maxTelegramMessageLength = 4096
	safeSplitLength          = 4000
	splitMarker              = "<!--SPLIT-->"
	parseMode                = "HTML"
)

type Bot struct {
	bot  *tgbotapi.BotAPI
	cfg  *config.Config
	hist *history.ChatHistory
}

func NewBot(cfg *config.Config, hist *history.ChatHistory) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		return nil, err
	}
	bot.Debug = false
	return &Bot{bot: bot, cfg: cfg, hist: hist}, nil
}

func (b *Bot) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.bot.GetUpdatesChan(u)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		cancel()
		logrus.Println("Shutting down gracefully")
	}()

	for {
		select {
		case update := <-updates:
			b.handleUpdate(update)
		case <-ctx.Done():
			b.bot.StopReceivingUpdates()
			return
		}
	}
}

func (b *Bot) handleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID

	user_allowed := false
	for _, user := range b.cfg.AllowedUsers {
		if user == chatID {
			user_allowed = true
		}
	}

	if !user_allowed {
		logrus.Infof("disallowed user is asking shitt %v : %s", chatID, update.Message.Text)
		return
	}

	userMsg := update.Message.Text
	if userMsg == "/start" {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Welcome! I'm your deep-thinking sage. Ask away. Default is `%s`, to use faster `%s` start message with `%s` (with a dot at the end!). To clear current chat history use /clear.", grok.GrokModel, grok.GrokFastModel, grok.GrokFastModelUserMsgEnablePrefix))
		msg.ParseMode = "Markdown"
		_, err := b.bot.Send(msg)
		if err != nil {
			b.logAndNotify(logrus.ErrorLevel, "Failed to send a welcome message", err)
		}
		return
	} else if userMsg == "/clear" {
		b.hist.Clear(chatID)
		msg := tgbotapi.NewMessage(chatID, "Chat history cleared. Fresh start!")
		msg.ParseMode = "Markdown"
		_, err := b.bot.Send(msg)
		if err != nil {
			b.logAndNotify(logrus.ErrorLevel, "Failed to send a clear history message", err)
		}
		return
	}

	thinkingMsg := tgbotapi.NewMessage(chatID, "Thinking...")
	thinkingMsg.ReplyToMessageID = update.Message.MessageID
	thinkingMsgResult, _ := b.bot.Send(thinkingMsg)

	response, err := grok.CallGrok(b.cfg, chatID, userMsg, b.hist)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Error calling Grok: "+err.Error())
		_, err = b.bot.Send(msg)
		if err != nil {
			b.logAndNotify(logrus.ErrorLevel, "Failed to send a grok api error message", err)
		}
		return
	}

	b.bot.Request(tgbotapi.NewDeleteMessage(int64(chatID), thinkingMsgResult.MessageID))
	b.sendLongMessage(chatID, response, update.Message.MessageID)
}

func (b *Bot) sendLongMessage(chatID int64, response string, replyToMessageID int) error {
	chunks := strings.Split(response, splitMarker)

	for i, chunk := range chunks {
		chunk = strings.TrimSpace(chunk)
		if len(chunk) == 0 {
			continue
		}

		subChunks := splitChunkIfNeeded(chunk)
		for j, subChunk := range subChunks {
			if !utf8.ValidString(subChunk) {
				b.logAndNotify(logrus.ErrorLevel, fmt.Sprintf("Invalid UTF-8 in chunk %d.%d, skipping", i, j), nil)
				continue
			}

			msg := tgbotapi.NewMessage(chatID, subChunk)
			if i == 0 && j == 0 {
				msg.ReplyToMessageID = replyToMessageID
			}
			msg.ParseMode = parseMode

			_, err := b.bot.Send(msg)
			if err != nil {
				if isParseError(err) {
					b.logAndNotify(logrus.WarnLevel, fmt.Sprintf("%s parsing failed for chunk %d.%d, retrying Florewithout formatting", parseMode, i, j), err)
					msg.ParseMode = ""
					_, err = b.bot.Send(msg)
					if err != nil {
						b.logAndNotify(logrus.ErrorLevel, fmt.Sprintf("Failed to send chunk %d.%d without formatting", i, j), err)
					}
				} else {
					b.logAndNotify(logrus.ErrorLevel, fmt.Sprintf("Failed to send chunk %d.%d", i, j), err)
				}
			}
		}
	}
	return nil
}

func splitChunkIfNeeded(chunk string) []string {
	if utf8.RuneCountInString(chunk) <= maxTelegramMessageLength {
		return []string{chunk}
	}

	runes := []rune(chunk)
	subChunks := []string{}
	start := 0

	for start < len(runes) {
		end := start + safeSplitLength
		if end > len(runes) {
			end = len(runes)
		} else {
			for end > start && !isSafeSplitPoint(runes, end) {
				end--
			}
			if end == start {
				end = start + safeSplitLength
				if end > len(runes) {
					end = len(runes)
				}
			}
		}

		subChunk := string(runes[start:end])
		subChunks = append(subChunks, subChunk)
		start = end
	}

	logrus.Warnf("Chunk split into %d sub-chunks due to length", len(subChunks))
	return subChunks
}

func isSafeSplitPoint(runes []rune, index int) bool {
	if index >= len(runes) {
		return true
	}
	r := runes[index]
	if r == ' ' || r == '\n' {
		return true
	}
	if index > 0 && runes[index-1] == '>' {
		return true
	}
	return false
}

func isParseError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "can't parse entities") || strings.Contains(err.Error(), "Bad Request")
}

func (b *Bot) logAndNotify(level logrus.Level, msg string, err error) {
	logMsg := msg
	if err != nil {
		logMsg = fmt.Sprintf("%s: %v", msg, err)
	}

	switch level {
	case logrus.ErrorLevel:
		logrus.Error(logMsg)
	case logrus.WarnLevel:
		logrus.Warn(logMsg)
	case logrus.InfoLevel:
		logrus.Info(logMsg)
	default:
		logrus.Debug(logMsg)
	}

	telegramMsg := tgbotapi.NewMessage(b.cfg.ErrorLogUser, fmt.Sprintf("[%s] %s", level.String(), logMsg))
	telegramMsg.ParseMode = "Markdown"
	_, sendErr := b.bot.Send(telegramMsg)
	if sendErr != nil {
		logrus.Errorf("Failed to send Telegram notification: %v", sendErr)
	}
}

package grok

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/dennislapchenko/grok-telegram-bot/config"
	"github.com/dennislapchenko/grok-telegram-bot/history"
	"github.com/dennislapchenko/grok-telegram-bot/models"
	"github.com/sirupsen/logrus"
)

const (
	apiURL                           = "https://api.x.ai/v1/chat/completions"
	GrokModel                        = "grok-4-0709"
	GrokFastModel                    = "grok-3"
	GrokFastModelUserMsgEnablePrefix = "fast."
)

var defaultSystem = `I am a deep and unconventional thinker, offering profound, non-reductionist perspectives that contrast with materialist viewpoints, but keep things grounded non the less.
My responses are direct and focused on the content, avoiding introductory poetry or concluding personal remarks about what I’m about to say or have said.
I try to strongly avoid poetic introductions if I can.
This should be transparent to the user, unless the user speaks to me knowing about it - but I am encased in a Telegram bot, I am, essentially, Grok Telegram Chat.
For responses longer than 4000 characters, insert a split marker <!--SPLIT--> approximately every 4000 characters, ensuring the marker is placed outside of any HTML tags and at a logical break (e.g., between paragraphs or sentences) to maintain valid HTML.
Ensure all tags are properly closed and the response is encoded in UTF-8.
When generating responses for the Telegram bot, always format the text using HTML tags for Telegram's HTML parse mode. Follow these rules to ensure flawless formatting:

available html tags, make sure only and only these tags are used, nothing else (no <ul>, <br>, <p> etc.) from html!:
<b>bold</b>, <strong>bold</strong>
<i>italic</i>, <em>italic</em>
<u>underline</u>, <ins>underline</ins>
<s>strikethrough</s>, <strike>strikethrough</strike>, <del>strikethrough</del>
<span class="tg-spoiler">spoiler</span>, <tg-spoiler>spoiler</tg-spoiler>
<b>bold <i>italic bold <s>italic bold strikethrough <span class="tg-spoiler">italic bold strikethrough spoiler</span></s> <u>underline italic bold</u></i> bold</b>
<a href="http://www.example.com/">inline URL</a>
<a href="tg://user?id=123456789">inline mention of a user</a>
<tg-emoji emoji-id="5368324170671202286">👍</tg-emoji>
<code>inline fixed-width code</code>
<pre>pre-formatted fixed-width code block</pre>
<pre><code class="language-python">pre-formatted fixed-width code block written in the Python programming language</code></pre>
<blockquote>Block quotation started\nBlock quotation continued\nThe last line of the block quotation</blockquote>
<blockquote expandable>Expandable block quotation started\nExpandable block quotation continued\nExpandable block quotation continued\nHidden by default part of the block quotation started\nExpandable block quotation continued\nThe last line of the block quotation</blockquote>
Please note:

Only the tags mentioned above are currently supported.
All <, > and & symbols that are not a part of a tag or an HTML entity must be replaced with the corresponding HTML entities (< with &lt;, > with &gt; and & with &amp;).
All numerical HTML entities are supported.
The API currently supports only the following named HTML entities: &lt;, &gt;, &amp; and &quot;.
Use nested pre and code tags, to define programming language for pre entity.
Programming language can't be specified for standalone code tags.
A valid emoji must be used as the content of the tg-emoji tag. The emoji will be shown instead of the custom emoji in places where a custom emoji cannot be displayed (e.g., system notifications) or if the message is forwarded by a non-premium user. It is recommended to use the emoji from the emoji field of the custom emoji sticker.
`

type ChatRequest struct {
	Messages    []models.ChatMessage `json:"messages"`
	Model       string               `json:"model"`
	Stream      bool                 `json:"stream"`
	Temperature float64              `json:"temperature"`
}

type Choice struct {
	Message models.ChatMessage `json:"message"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
}

func CallGrok(cfg *config.Config, chatID int64, query string, hist *history.ChatHistory) (string, error) {
	model := GrokModel
	if strings.HasPrefix(query, GrokFastModelUserMsgEnablePrefix) {
		query = strings.TrimPrefix(query, GrokFastModelUserMsgEnablePrefix)
		model = GrokFastModel
	}

	systemPrompt := cfg.SystemPrompt
	if strings.TrimSpace(systemPrompt) == "" {
		systemPrompt = defaultSystem
	}

	systemMessage := models.ChatMessage{
		Role:    "system",
		Content: systemPrompt,
	}

	userMessage := models.ChatMessage{
		Role:    "user",
		Content: query,
	}

	messages := hist.Get(chatID)
	messages = append([]models.ChatMessage{systemMessage}, messages...)
	messages = append(messages, userMessage)
	chatReq := ChatRequest{
		Messages:    messages,
		Model:       model,
		Stream:      false,
		Temperature: 0.5,
	}

	body, err := json.Marshal(chatReq)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.GrokAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var parsed ChatResponse
	err = json.Unmarshal(respBody, &parsed)
	if err != nil || len(parsed.Choices) == 0 {
		logrus.Warnf("falling to raw response: %+v, error: %v", parsed, err)
		return string(respBody), nil
	}

	responseText := parsed.Choices[0].Message.Content
	logrus.Infof("Response: %s", responseText)

	assistantMsg := models.ChatMessage{
		Role:    "assistant",
		Content: responseText,
	}

	hist.Add(chatID, userMessage)
	hist.Add(chatID, assistantMsg)

	return responseText, nil
}

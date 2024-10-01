package messagesender

import (
	"fmt"
	"net/http"
	"time"

	queue "github.com/kupalovmuhammadjon/Queue"
)

type TelegramBot interface {
	Send(message string) error
}

type telegramBot struct {
	messageQueue queue.Queue
	cfg          Options
}

type Options struct {
	BotToken string
	ChatIds  []string
}

func New(config Options) (TelegramBot, error) {
	if config.BotToken == "" {
		return nil, fmt.Errorf("telegram bot token is required")
    }

	if len(config.ChatIds) == 0 {
        return nil, fmt.Errorf("at least one chat ID is required")
    }

	model := &telegramBot{
		messageQueue: queue.NewQueue(),
		cfg:          config,
	}

	go model.MakeRequestsWithRPSLimit(50)
	
	return model, nil

}

func (t *telegramBot) Send(message string) error {

	for _, e := range t.cfg.ChatIds {
		url := fmt.Sprintf("https://api.telegram.org/bot"+t.cfg.BotToken+"/sendMessage?chat_id="+e+"&text=%s", message)
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		t.messageQueue.Push(request)
	}

	return nil
}

func (t *telegramBot) MakeRequestsWithRPSLimit(rps int) {
	ticker := time.NewTicker(time.Second / time.Duration(rps))
	defer ticker.Stop()

	client := &http.Client{}

	for {
		select {
		case <-ticker.C:
			if t.messageQueue.IsEmpty() {
				continue
			}

			req := t.messageQueue.Pop().(*http.Request)
			go func(r *http.Request) {
				resp, err := client.Do(r)
				if err != nil {
					fmt.Printf("Error sending message: %v\n", err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					fmt.Printf("Unexpected status code: %d\n", resp.StatusCode)
				}
			}(req)
		}
	}
}

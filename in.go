package messagesender

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	queue "github.com/kupalovmuhammadjon/Queue"
)

type TelegramBot interface {
	Send(message string) error
	Wait()
}

type telegramBot struct {
	messageQueue queue.Queue
	cfg          Options
	wg           sync.WaitGroup // WaitGroup to track goroutines
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

	go model.makeRequestsWithRPSLimit(20)

	return model, nil
}

func (t *telegramBot) Send(message string) error {
	for _, chatId := range t.cfg.ChatIds {
		fmt.Println("request added to queue: ", chatId)
		url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s", t.cfg.BotToken, chatId, message)
		request, err := http.NewRequest("POST", url, nil)
		if err != nil {
			return err
		}

		// Increment the WaitGroup counter before starting the goroutine
		t.wg.Add(1)
		t.messageQueue.Push(request)
	}

	return nil
}

func (t *telegramBot) makeRequestsWithRPSLimit(rps int) {
	ticker := time.NewTicker(time.Minute / time.Duration(rps))
	defer ticker.Stop()

	client := &http.Client{}

	for {
		select {
		case <-ticker.C:
			if t.messageQueue.IsEmpty() {
				continue
			}

			req, ok := t.messageQueue.Pop().(*http.Request)
			if !ok {
				fmt.Println("Error: unexpected type in message queue")
				continue
			}

			go func(req *http.Request) {
				defer t.wg.Done()
				resp, err := client.Do(req)
				if err != nil {
					fmt.Printf("Error sending message: %v\n", err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					fmt.Printf("Unexpected status code: %d\n", resp.StatusCode)
				}
				fmt.Println("request done: ")
			}(req)
		}
	}
}

func (t *telegramBot) Wait() {
	t.wg.Wait() // Wait for all the goroutines to finish
}

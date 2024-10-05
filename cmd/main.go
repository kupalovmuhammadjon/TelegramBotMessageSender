package main

import (
	"fmt"
	messagesender "telegrambotmessagesender"
)

func main() {
	bot, err := messagesender.New(messagesender.Options{
		BotToken: "7702142395:AAEis3BuRQ-QOCbmvyxSDGrFxI4Ie2EY2qw",
		ChatIds:  []string{"-1002296065689"},
	})
	if err != nil {
		panic(err)
	}
	for i := 0; i < 50; i++ {
		// go func() {
			err = bot.Send(fmt.Sprintf("%d", i))
			if err != nil {
				panic(err)
			}
		// }()

	}

	bot.Wait()
}

package main

import (
	"fmt"
	"log"
	"os"

	rivescript "github.com/aichaos/rivescript-go"
)

func main() {
	userID := os.Args[1]
	message := os.Args[2]

	bot := rivescript.New(rivescript.WithUTF8())
	err := bot.LoadFile("./script.rive")
	if err != nil {
		log.Fatalf("Error loading from file: %s\n", err)
	}
	bot.SortReplies()
	reply, _ := bot.Reply(userID, message)
	fmt.Println(reply)
}

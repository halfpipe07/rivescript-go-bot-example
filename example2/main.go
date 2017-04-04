package main

import (
	"fmt"
	"log"
	"os"

	rivescript "github.com/aichaos/rivescript-go"
)

var bot *rivescript.RiveScript

func loadBot() {
	bot = rivescript.New(rivescript.WithUTF8())
	err := bot.LoadFile("./script.rive")
	if err != nil {
		log.Fatalf("Error loading from file: %s\n", err)
	}
	bot.SortReplies()
	fmt.Printf("Bot loaded...\n> ")
}

func main() {
	userID := os.Args[1]
	loadBot()
	var message string
	for {
		_, _ = fmt.Scanf("%s", &message)
		reply := getReply(message, userID)
		fmt.Printf("%s\n> ", reply)
	}
}

func getReply(message string, userID string) string {
	reply, _ := bot.Reply(userID, message)
	fmt.Println(reply)
	return ""

}

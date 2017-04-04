package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"./nlp"

	"github.com/aichaos/rivescript-go"
	rss "github.com/aichaos/rivescript-go/src"
)

const (
	apiAIAPIKey    = "YOUR_APIAI_KEY"
	apiAISessionID = "393933939393939333l"
	fBToken        = "123456"
	fbPageToken    = "YOUR_FB_PAGE_TOKEN"
)

var bot *BotBrain

func handler(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		handleChallenge(rw, req)
	} else {
		handleMessage(rw, req)
	}
}

func handleChallenge(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("Request:\n%+v\n", *req)
	u, _ := url.Parse(req.RequestURI)
	q := u.Query()
	challenge := q.Get("hub.challenge")
	token := q.Get("hub.verify_token")
	if token == fBToken {
		fmt.Fprintf(rw, challenge)
		return
	}
	fmt.Fprintf(rw, "FAIL!")
}

type FBMessage struct {
	Sender struct {
		ID string `json:"id"`
	} `json:"sender"`
	Recipient struct {
		ID string `json:"id"`
	} `json:"recipient"`
	Timestamp int64 `json:"timestamp"`
	Message   struct {
		Mid  string `json:"mid"`
		Seq  int    `json:"seq"`
		Text string `json:"text"`
	} `json:"message"`
}

type FBRecipient struct {
}
type FBResponse struct {
	Recipient struct {
		ID string `json:"id"`
	} `json:"recipient"`
	Message struct {
		Text string `json:"text"`
	} `json:"message"`
}

func NewResponse(recipient string, message string) FBResponse {
	res := FBResponse{}
	res.Recipient.ID = recipient
	res.Message.Text = message
	return res
}

type FBMessaging struct {
	Object string `json:"object"`
	Entry  []struct {
		ID        string      `json:"id"`
		Time      int64       `json:"time"`
		Messaging []FBMessage `json:"messaging"`
	} `json:"entry"`
}

func handleMessage(rw http.ResponseWriter, req *http.Request) {
	b := req.Body
	defer b.Close()
	stuff, _ := ioutil.ReadAll(b)
	var msging FBMessaging
	err := json.Unmarshal(stuff, &msging)
	if err != nil {
		log.Printf("Error with JSON: %s\n", err)
		fmt.Fprintf(rw, "NO!!!\n")
		return
	}
	for _, entry := range msging.Entry {
		for _, msg := range entry.Messaging {
			text := msg.Message.Text
			reply, _ := bot.Reply(msg.Sender.ID, text)
			response := NewResponse(msg.Sender.ID, reply)
			data, _ := json.Marshal(response)
			r := bytes.NewReader(data)
			uri := "https://graph.facebook.com/v2.6/me/messages"
			req, err := http.NewRequest("POST", uri, r)
			if err != nil {
				log.Printf("Error with post request: %s\n", err)
				return
			}
			q := req.URL.Query()
			q.Add("access_token", fbPageToken)
			req.URL.RawQuery = q.Encode()
			req.Header.Add("Content-Type", "application/json")
			cli := &http.Client{}
			cli.Do(req)
		}
	}

	fmt.Fprintf(rw, "OK\n")
}

func main() {
	bot = New()
	http.HandleFunc("/providers/facebook/webhook", handler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Serving requests on port %s\n", port)
	http.ListenAndServe(":"+port, nil)
	// username := os.Args[1]
	// message := os.Args[2]
	// Get a reply.
	// reply, _ := bot.Reply(username, message)
	// args := []string{message}
	// reply := getIntent(args)
	// fmt.Println(reply)
}

type BotBrain struct {
	bot *rivescript.RiveScript
}

func New() *BotBrain {
	// // Create a new bot with the default settings.
	// bb.bot := rivescript.New(nil)
	brain := BotBrain{}
	bot := rivescript.New(rivescript.WithUTF8())

	bot.SetSubroutine("getIntent", getIntent)

	// Load a directory full of RiveScript documents (.rive files)
	// err := bot.LoadDirectory("eg/brain")
	// if err != nil {
	// 	fmt.Printf("Error loading from directory: %s\n", err)
	// }

	// Load an individual file.
	err := bot.LoadFile("./rivescripts/burgers.rive")
	if err != nil {
		fmt.Printf("Error loading from file: %s\n", err)
	}

	// Sort the replies after loading them!
	bot.SortReplies()

	brain.bot = bot

	return &brain
}

func (brain *BotBrain) isIntent(text string) bool {
	matched, err := regexp.MatchString("^intent ", text)
	if err != nil {
		return false
	}
	return matched
}

func (brain *BotBrain) Reply(userID string, message string) (string, error) {
	reply, err := brain.bot.Reply(userID, message)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	if brain.isIntent(reply) {
		return brain.Reply(userID, reply)
	}

	return reply, nil
	// answer, _ := brain.parseAnswer(reply)
	// return answer, nil
}

func getIntent(rs *rss.RiveScript, messages []string) string {
	cli := nlp.NewClient(apiAIAPIKey)
	res, err := cli.GetReply(messages[0], apiAISessionID)
	if err != nil {
		panic(err)
	}
	if len(res.Parameters) != 0 {
		fmt.Println(res.Parameters)
	}
	intentName := strings.Replace(res.Name, "_", " ", -1)
	return "intent " + intentName
}

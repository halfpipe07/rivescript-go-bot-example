package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aichaos/rivescript-go"
)

func doSomething(args []string) string {
	return "Something!"
}

var apiAIAPIKey = "2d8c9a781c2a450fa2598424a894c4f0"
var apiAISessionID = "393933939393939333l"

var funcs = map[string]func([]string) string{
	"printSomething": doSomething,
}

var bot *BotBrain

func main() {
	message := os.Args[1]
	fmt.Println(message)
	// Get a reply.
	// // bot = New()
	// reply, _ := bot.Reply("local-user", message)
	args := []string{message}
	reply := getIntent(args)
	fmt.Println(reply)
}

type BotBrain struct {
	bot *rivescript.RiveScript
}

func New() *BotBrain {
	// // Create a new bot with the default settings.
	// bb.bot := rivescript.New(nil)
	brain := BotBrain{}

	// To enable UTF-8 mode, you'd have initialized the bot like:
	brain.bot = rivescript.New(rivescript.WithUTF8())

	// Load a directory full of RiveScript documents (.rive files)
	// err := bot.LoadDirectory("eg/brain")
	// if err != nil {
	// 	fmt.Printf("Error loading from directory: %s\n", err)
	// }

	// Load an individual file.
	err := brain.bot.LoadFile("./rivescripts/burgers.rive")
	if err != nil {
		fmt.Printf("Error loading from file: %s\n", err)
	}

	// Sort the replies after loading them!
	brain.bot.SortReplies()

	return &brain
}

func (brain *BotBrain) Reply(userID string, message string) (string, error) {
	reply, err := brain.bot.Reply(userID, message)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	} else {
		fmt.Printf("The bot says: %s\n", reply)
	}

	answer, _ := brain.parseAnswer(reply)
	return answer, nil
}

func (brain *BotBrain) parseAnswer(message string) (string, error) {
	r, _ := regexp.Compile("^!code (.+)")
	isMatch := r.MatchString(message)
	if isMatch {
		matches := r.FindStringSubmatch(message)
		if len(matches) < 2 {
			return "", errors.New("Message format incorrect, couldn't figure out message.")
		}
		parts := strings.Split(matches[1], "|")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return brain.call(parts), nil
	}
	return message, nil
}

func (brain *BotBrain) call(args []string) string {
	fName := args[0]
	fArgs := args[1:len(args)]
	return funcs[fName](fArgs)
}

type APIAIResponse struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Result    struct {
		Source           string            `json:"source"`
		ResolvedQuery    string            `json:"resolvedQuery"`
		Action           string            `json:"action"`
		ActionIncomplete bool              `json:"actionIncomplete"`
		Parameters       map[string]string `json:"parameters"`
		Contexts         []struct {
			Name       string `json:"name"`
			Parameters struct {
				Name string `json:"name"`
			} `json:"parameters"`
			Lifespan int `json:"lifespan"`
		} `json:"contexts"`
		Metadata struct {
			IntentID   string `json:"intentId"`
			IntentName string `json:"intentName"`
		} `json:"metadata"`
		Fulfillment struct {
			Speech string `json:"speech"`
		} `json:"fulfillment"`
	} `json:"result"`
	Status struct {
		Code      int    `json:"code"`
		ErrorType string `json:"errorType"`
	} `json:"status"`
}

type Intent struct {
	Name       string
	Parameters map[string]string
}

type APIAIClient struct {
	cli      *http.Client
	apiToken string
}

func NewClient(apiToken string) *APIAIClient {
	cli := APIAIClient{}
	cli.cli = &http.Client{}
	cli.apiToken = apiToken
	return &cli
}

func (cli *APIAIClient) GetReply(message string) (*Intent, error) {
	uri := "https://api.api.ai/api/query?v=20150910&lang=en&timezone=2017-03-30T14:11:36+0200&query=%s&sessionId=%s"
	m := url.QueryEscape(message)
	sid := url.QueryEscape(apiAISessionID)
	finalURI := fmt.Sprintf(uri, m, sid)
	req, err := http.NewRequest("GET", finalURI, nil)
	if err != nil {
		return nil, err
	}
	// Replace 9ea93023b7274cfbb392b289658cff0b by your Client access token
	req.Header.Add("Authorization", "Bearer "+cli.apiToken)

	resp, err := cli.cli.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// Fill the record with the data from the JSON
	var record APIAIResponse

	// Use json.Decode for reading streams of JSON data
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		return nil, err
	}
	intent := Intent{
		Name:       record.Result.Metadata.IntentName,
		Parameters: record.Result.Parameters,
	}
	return &intent, nil
}

func getIntent(args []string) string {
	cli := NewClient(apiAIAPIKey)
	res, err := cli.GetReply(args[0])
	if err != nil {
		panic(err)
	}
	fmt.Println(res.Parameters)
	return res.Name
}

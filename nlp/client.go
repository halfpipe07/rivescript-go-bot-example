package nlp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

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

func (cli *APIAIClient) GetReply(message string, session string) (*Intent, error) {
	uri := "https://api.api.ai/api/query?v=20150910&lang=en&timezone=2017-03-30T14:11:36+0200&query=%s&sessionId=%s"
	m := url.QueryEscape(message)
	sid := url.QueryEscape(session)
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

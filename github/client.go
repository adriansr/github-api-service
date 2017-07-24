package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"io/ioutil"

	base "github.com/adriansr/github-api-service"
	"github.com/adriansr/github-api-service/util"
)

/* Client ... */
type Client struct {
	username, password string
	// clients are safe for concurrent use
	httpClient http.Client
	//request50, request100, request150page2 http.Request
}

type GitHubSearchResponse struct {
	TotalCount int  `json:"total_count"`
	Incomplete bool `json:"incomplete_results"`
	Items      []base.User
}

const (
	userAgent = "adriansr/github-api-service:1.0"
	debugBody = false
)

func NewClient(username, password string, timeout time.Duration) (base.Client, error) {
	return &Client{
		username:   username,
		password:   password,
		httpClient: http.Client{Timeout: timeout},
	}, nil
}

func (client *Client) Search(location string, count int) ([]base.User, error) {
	if count != 50 && count != 100 && count != 150 {
		return nil, util.NewError("count parameter out of range")
	}

	limit := util.Min(count, 100)
	result, err := client.singleRequest(location, limit, 1)
	if err != nil {
		return nil, err
	}
	if len(result.Items) < limit || limit == count {
		return result.Items, nil
	}
	result2, err := client.singleRequest(location, 50, 3)
	if err != nil {
		return nil, err
	}
	return append(result.Items, result2.Items...), nil
}

func (client *Client) singleRequest(location string, count int, page int) (*GitHubSearchResponse, error) {
	url := fmt.Sprintf("https://api.github.com/search/users?sort=repositories&order=desc&per_page=%d&page=%d&q=location:%s",
		count, page, location)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, util.WrapError("failed creating a request object", err)
	}
	if len(client.username) > 0 && len(client.password) > 0 {
		request.SetBasicAuth(client.username, client.password)
	}
	request.Header.Add("User-Agent", userAgent)
	response, err := client.httpClient.Do(request)
	if err != nil {
		return nil, util.WrapError("failed creating an HTTP client", err)
	}
	defer response.Body.Close()
	var searchResult GitHubSearchResponse
	if debugBody {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, util.WrapError("failed reading response body", err)
		}
		fmt.Printf("Received body [%d bytes] <<<%s>>>", len(body), body)
		if err := json.Unmarshal(body, &searchResult); err != nil {
			return nil, util.WrapError("Failed unmarshalling json response", err)
		}
	} else {
		if err := json.NewDecoder(response.Body).Decode(&searchResult); err != nil {
			return nil, util.WrapError("failed decoding json response", err)
		}
	}
	return &searchResult, nil
}

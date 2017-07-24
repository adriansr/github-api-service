package githubapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"io/ioutil"

	"github.com/adriansr/github-api-service/model"
	"github.com/adriansr/github-api-service/util"
)

// Client encapsulates the fields required to perform queries to the GitHub
// search API
type Client struct {
	username, password string
	apiUrl             string
	// clients are safe for concurrent use
	httpClient http.Client
}

// (private) representation of a github user as returned by the search API,
// featuring only the required fields
type githubUser struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
}

// (private) representation of a search API response
type searchResponse struct {
	TotalCount int          `json:"total_count"`
	Incomplete bool         `json:"incomplete_results"`
	Items      []githubUser `json:"items"`
}

const (
	// user-agent for http client side, using the path to the source repository
	// as sugested in GitHub docs
	userAgent = "adriansr/github-api-service"
	debugBody = false
)

// NewClient returns a newly created Client to the GitHub API
func NewClient(username, password, apiUrl string, timeout time.Duration) (*Client, error) {
	return &Client{
		username:   username,
		password:   password,
		apiUrl:     apiUrl,
		httpClient: http.Client{Timeout: timeout},
	}, nil
}

// GetTopContributors queries the GitHub API for the `count` top contributors
// on the given location.
func (client *Client) GetTopContributors(location string, count int) ([]model.User, error) {
	if count != 50 && count != 100 && count != 150 {
		return nil, util.NewError("count parameter out of range")
	}

	// request up to 100 results, as GitHub search API currently limits to
	// 100 results per page
	// TODO: support an arbitrary limit
	limit := util.Min(count, 100)
	result, err := client.searchUsers(location, limit, 1)
	if err != nil {
		return nil, err
	}
	if len(result.Items) < limit || limit == count {
		return result.users(), nil
	}
	result2, err := client.searchUsers(location, 50, 3)
	if err != nil {
		return nil, err
	}
	return append(result.users(), result2.users()...), nil
}

// (private) transforms the internal representation of the list of
// users returned by a search query to the expected User. This is
// necessary to not be forced to use the same field names in the output
// json as GitHub uses. Alternatively a custom json-marshaler can be used.
func (response *searchResponse) users() []model.User {
	result := make([]model.User, len(response.Items))
	for idx, user := range response.Items {
		result[idx].ID = user.ID
		result[idx].Username = user.Login
	}
	return result
}

// (private) searchUsers perform a user search query against GitHub API
// filtering by location
func (client *Client) searchUsers(location string, count int, page int) (*searchResponse, error) {
	url := fmt.Sprintf("%s/search/users?sort=repositories&order=desc&per_page=%d&page=%d&q=location:%s",
		client.apiUrl, count, page, location)
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
	var searchResult searchResponse
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

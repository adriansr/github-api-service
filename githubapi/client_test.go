package githubapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/adriansr/github-api-service/model"
)

const (
	noUser      = ""
	noPass      = ""
	timeout     = 3 * time.Second
	partialJson = `{"total_count": 5, "items": [ { "id":1, }]}`
)

type RequestResponseTester struct {
	Request  *http.Request
	Code     int
	Response []byte
}

func (tester *RequestResponseTester) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	tester.Request = request
	writer.WriteHeader(tester.Code)
	writer.Write(tester.Response)
}

func toJSON(t *testing.T, obj searchResponse) []byte {
	if result, err := json.Marshal(obj); err == nil {
		return result
	} else {
		t.Fatal(err)
		return nil
	}
}

func assertEquals(t *testing.T, a []githubUser, b []model.User) {
	if len(a) != len(b) {
		t.Fatalf("Comparing %v vs %v : different sizes", a, b)
	}
	for i := 0; i < len(a); i++ {
		if a[i].ID != b[i].ID {
			t.Fatalf("Comparing %v vs %v : different ID", a[i], b[i])
		}
		if a[i].Login != b[i].Username {
			t.Fatalf("Comparing %v vs %v : different Name", a[i], b[i])
		}

	}
}

func makeResponse(totalCount int, incomplete bool, count int) searchResponse {
	items := make([]githubUser, count)
	for i := 0; i < count; i++ {
		items[i] = githubUser{ID: int64(i), Login: fmt.Sprintf("user_%d", i)}
	}
	return searchResponse{totalCount, incomplete, items}
}

func TestClient(t *testing.T) {
	response := makeResponse(5000, false, 50)
	handler := &RequestResponseTester{nil, 200, toJSON(t, response)}
	server := httptest.NewServer(handler)
	defer server.Close()

	client, err := NewClient(noUser, noPass, server.URL, timeout)
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.GetTopContributors("Barcelona", 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 50 {
		t.Fatalf("result: %v", result)
	}
	assertEquals(t, response.Items, result)
}

/*
	// test default count

	// test count

	// test count invalid

	// test city encoding
*/

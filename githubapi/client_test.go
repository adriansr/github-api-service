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
	noUser  = ""
	noPass  = ""
	timeout = 3 * time.Second
)

type RequestResponseTester struct {
	Request  *http.Request
	Code     int
	Response []byte
}

type RedirectHandler struct {
	destination string
}

func (tester *RequestResponseTester) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	tester.Request = request
	writer.WriteHeader(tester.Code)
	writer.Write(tester.Response)
}

func (handler RedirectHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Location", handler.destination)
	writer.WriteHeader(http.StatusTemporaryRedirect)
	writer.Write([]byte{})
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

func TestParsing(t *testing.T) {
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

func TestNoAuth(t *testing.T) {
	handler := &RequestResponseTester{nil, 500, []byte("bye")}
	server := httptest.NewServer(handler)
	defer server.Close()

	client, err := NewClient(noUser, noPass, server.URL, timeout)
	if err != nil {
		t.Fatal(err)
	}

	client.GetTopContributors("Barcelona", 50)

	user, pass, ok := handler.Request.BasicAuth()
	if user != "" || pass != "" || ok {
		t.Fatalf("Wrong auth: user:'%s' pass:'%s' valid:%v", user, pass, ok)
	}
}

func TestWithAuth(t *testing.T) {
	someUser, somePass := "someUser", "somePass"
	handler := &RequestResponseTester{nil, 500, []byte("bye")}
	server := httptest.NewServer(handler)
	defer server.Close()

	client, err := NewClient(someUser, somePass, server.URL, timeout)
	if err != nil {
		t.Fatal(err)
	}

	client.GetTopContributors("Barcelona", 50)

	user, pass, ok := handler.Request.BasicAuth()
	if user != someUser || pass != somePass || !ok {
		t.Fatalf("Wrong auth: user:'%s' pass:'%s' valid:%v", user, pass, ok)
	}
}

func TestEscaping(t *testing.T) {
	handler := &RequestResponseTester{nil, 500, []byte("bye")}
	server := httptest.NewServer(handler)
	defer server.Close()

	client, err := NewClient(noUser, noPass, server.URL, timeout)
	if err != nil {
		t.Fatal(err)
	}

	city := "Rio de Janeiro"
	client.GetTopContributors(city, 50)

	if handler.Request == nil {
		t.Fatal("request not sent")
	}
	expected := fmt.Sprintf("location:%s", city)
	query := handler.Request.URL.Query().Get("q")
	if query != expected {
		t.Fatalf("unexpected query string: '%s' vs '%s'", query, expected)
	}
}

func TestApiError(t *testing.T) {
	handler := &RequestResponseTester{nil, 500, []byte("bye")}
	server := httptest.NewServer(handler)
	defer server.Close()

	client, err := NewClient(noUser, noPass, server.URL, timeout)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.GetTopContributors("Barcelona", 50)
	if err == nil {
		t.Fatal("failure expected")
	}
}

func TestRedirect(t *testing.T) {
	response := makeResponse(5000, false, 50)
	handler := &RequestResponseTester{nil, 200, toJSON(t, response)}
	server := httptest.NewServer(handler)
	defer server.Close()

	server2 := httptest.NewServer(RedirectHandler{server.URL})
	client, err := NewClient(noUser, noPass, server2.URL, timeout)
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

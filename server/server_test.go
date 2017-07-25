package server

import (
	"fmt"
	"net/http"
	"testing"

	"time"

	"github.com/adriansr/github-api-service/model"
	"github.com/adriansr/github-api-service/util"
)

// Recorder helper to record calls to the TopContributorGetter interface
type Recorder struct {
	Users []model.User
	Error error
	City  string
	Count int
	Calls int
}

func newRecorder(countUsers int, err error) *Recorder {
	users := make([]model.User, countUsers)
	for i := 0; i < countUsers; i++ {
		users[i] = model.User{ID: int64(i), Username: fmt.Sprintf("user_%d", i)}
	}
	return &Recorder{Users: users, Error: err, Calls: 0}
}

func (recorder *Recorder) GetTopContributors(location string, count int) ([]model.User, error) {
	recorder.Calls++
	recorder.City = location
	recorder.Count = count
	return recorder.Users, recorder.Error
}

// ServerContext helper to create a testable HTTP server
type ServerContext struct {
	server     *Server
	terminator chan error
	t          *testing.T
}

func createServer(t *testing.T, handler model.TopContributorGetter) *ServerContext {
	server, err := New(":0", handler)
	if err != nil {
		t.Fatalf("failed creating server: %s", err)
	}

	ctx := &ServerContext{server, make(chan error, 1), t}

	go func() {
		err = server.Start()
		if err != nil {
			ctx.terminator <- err
		}
	}()

	ctx.awaitAlive()
	return ctx
}

func (ctx *ServerContext) pollStartError() error {
	select {
	case err := <-ctx.terminator:
		return err
	default:
		return nil
	}
}

func (ctx *ServerContext) stop() {
	if err := ctx.pollStartError(); err != nil {
		ctx.t.Fatalf("Server failed to start: %s", err)
	}
	if err := ctx.server.Stop(); err != nil {
		ctx.t.Fatalf("Server failed to stop: %s", err)
	}
}

func (ctx *ServerContext) url() string {
	return fmt.Sprintf("http://%s", ctx.server.Address.Addr().String())
}

func (ctx *ServerContext) awaitAlive() {
	const MaxAttempts = 5
	client := http.Client{Timeout: time.Second}
	url := fmt.Sprintf("%s/check-alive", ctx.url())

	for attempt := 0; attempt < MaxAttempts; attempt++ {
		if err := ctx.pollStartError(); err != nil {
			ctx.t.Fatalf("Server failed to start: %s", err)
		}
		resp, err := client.Head(url)
		if err == nil && resp.StatusCode == 404 {
			// server is alive and responding
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	ctx.t.Fatal("Server didn't start on time")
}

func TestServerStartStop(t *testing.T) {
	server := createServer(t, nil)
	server.stop()
}

func TestServerValidRequest(t *testing.T) {
	recorder := newRecorder(10, nil)
	server := createServer(t, recorder)

	client := http.Client{Timeout: time.Second}
	url := fmt.Sprintf("%s/api/top-contributors?city=CITY&count=100", server.url())

	response, err := client.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 200 {
		t.Fatalf("got HTTP code %d", response.StatusCode)
	}
	if recorder.Calls != 1 {
		t.Fatalf("one query expected, got %d", recorder.Calls)
	}

	if recorder.City != "CITY" {
		t.Fatalf("wrong city, got %s", recorder.City)
	}
	if recorder.Count != 100 {
		t.Fatalf("wrong count, got %d", recorder.Count)
	}
	server.stop()
}

func TestServerDefaultCount(t *testing.T) {
	recorder := newRecorder(10, nil)
	server := createServer(t, recorder)

	client := http.Client{Timeout: time.Second}
	url := fmt.Sprintf("%s/api/top-contributors?city=CITY", server.url())

	response, err := client.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 200 {
		t.Fatalf("got HTTP code %d", response.StatusCode)
	}
	if recorder.Calls != 1 {
		t.Fatalf("one query expected, got %d", recorder.Calls)
	}

	if recorder.City != "CITY" {
		t.Fatalf("wrong city, got %s", recorder.City)
	}
	if recorder.Count != 50 {
		t.Fatalf("wrong count, got %d", recorder.Count)
	}
	server.stop()
}

func TestServerInvalidCount(t *testing.T) {
	recorder := newRecorder(10, nil)
	server := createServer(t, recorder)

	client := http.Client{Timeout: time.Second}
	url := fmt.Sprintf("%s/api/top-contributors?city=CITY&count=333", server.url())

	response, err := client.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 400 {
		t.Fatalf("got HTTP code %d", response.StatusCode)
	}
	if recorder.Calls != 0 {
		t.Fatalf("no query expected, got %d", recorder.Calls)
	}

	server.stop()
}

func TestServerNoCity(t *testing.T) {
	recorder := newRecorder(10, nil)
	server := createServer(t, recorder)

	client := http.Client{Timeout: time.Second}
	url := fmt.Sprintf("%s/api/top-contributors", server.url())

	response, err := client.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 400 {
		t.Fatalf("got HTTP code %d", response.StatusCode)
	}

	if recorder.Calls != 0 {
		t.Fatalf("no query expected, got %d", recorder.Calls)
	}

	server.stop()
}

func TestServerUrlEncoded(t *testing.T) {
	recorder := newRecorder(10, nil)
	server := createServer(t, recorder)

	client := http.Client{Timeout: time.Second}
	url := fmt.Sprintf("%s/api/top-contributors?city=Sao%%20Paulo", server.url())

	response, err := client.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 200 {
		t.Fatalf("got HTTP code %d", response.StatusCode)
	}
	if recorder.Calls != 1 {
		t.Fatalf("one query expected, got %d", recorder.Calls)
	}
	if recorder.City != "Sao Paulo" {
		t.Fatalf("wrong city, got %s", recorder.City)
	}
	server.stop()
}

func TestQueryFailure(t *testing.T) {
	recorder := newRecorder(0, util.NewError("error"))
	server := createServer(t, recorder)

	client := http.Client{Timeout: time.Second}
	url := fmt.Sprintf("%s/api/top-contributors?city=Sao%%20Paulo", server.url())

	response, err := client.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 500 {
		t.Fatalf("got HTTP code %d", response.StatusCode)
	}
	if recorder.Calls != 1 {
		t.Fatalf("one query expected, got %d", recorder.Calls)
	}
	server.stop()
}

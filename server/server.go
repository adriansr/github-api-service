package server

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/adriansr/github-api-service/model"
	"github.com/adriansr/github-api-service/util"
)

// Server struct contains the fields to model our API server
type Server struct {
	// Address the server is bound to
	Address net.Listener

	// interface to fetch the top contributors
	client model.TopContributorGetter

	// multiplexor for requests
	handler *http.ServeMux

	// internal server handle
	underlying *http.Server
}

// ApiError struct is used to represent the error responses from the API
type ApiError struct {
	Error string `json:"error"`
}

const (
	// path for the API endpoint
	apiPath = "/api/top-contributors"
	// `Server` header to use in HTTP responses
	serverName = "adriansr/github-api-service"
	// by default 50 results are fetched if count is not specified
	defaultCount = 50
)

func setCommonHeaders(writer http.ResponseWriter) {
	writer.Header().Add("Content-Type", "application/json")
	writer.Header().Add("Server", serverName)
}

func sendError(writer http.ResponseWriter, code int, msg string) {
	object := ApiError{msg}
	body, err := json.Marshal(object)
	if err != nil {
		code = http.StatusInternalServerError
		body = []byte(`{"error": "internal error"}`)
	}
	writer.WriteHeader(code)
	writer.Write(body)
	log.Printf("Error response %d '%s'", code, msg)
}

// ServeHTTP handles HTTP requests to the API endpoint
func (server *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	setCommonHeaders(writer)

	// only accept GET requests
	if request.Method != "GET" {
		// required when sending 405 Method Not Allowed
		writer.Header().Add("Allow", "GET")
		sendError(writer, http.StatusMethodNotAllowed, "only GET requests allowed")
		return
	}

	// check count
	params := request.URL.Query()
	count, err := strconv.Atoi(params.Get("count"))
	if err != nil {
		count = defaultCount
	}

	if count != 50 && count != 100 && count != 150 {
		sendError(writer, http.StatusBadRequest, "count parameter not valid")
		return
	}

	// check city
	city := params.Get("city")
	if len(city) == 0 {
		sendError(writer, http.StatusBadRequest, "missing parameter: city")
		return
	}

	// forward request to the TopContributorGetter instace
	result, err := server.client.GetTopContributors(city, count)
	if err != nil {
		sendError(writer, http.StatusInternalServerError,
			"query failed: "+err.Error())
		return
	}

	// convert to JSON
	body, err := json.Marshal(result)
	if err != nil {
		sendError(writer, http.StatusInternalServerError,
			"output representation failed: "+err.Error())
		return
	}
	writer.WriteHeader(http.StatusOK)
	writer.Write(body)
	log.Printf("Processed request (%d results)", len(result))
}

func notFound(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Request for unknown path: %s", request.URL)
	http.NotFound(writer, request)
}

// New creates an API Server that binds to the given address and uses the
// passed client to get the top contributors
func New(address string, client model.TopContributorGetter) (*Server, error) {

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, util.WrapError("Listen failed", err)
	}
	server := &Server{listener, client, http.NewServeMux(), nil}
	server.handler.Handle(apiPath, server)
	// attach a NotFound handler to / so it can log 404 errors
	server.handler.HandleFunc("/", notFound)
	log.Printf("Registered API endpoint '%s'", apiPath)
	return server, nil
}

// Start starts accepting requests on the server, blocking until explicitly
// terminated
func (server *Server) Start() error {
	if server.underlying != nil {
		return util.NewError("already running")
	}
	log.Printf("Accepting requests at %s", server.Address.Addr())
	server.underlying = &http.Server{Handler: server.handler}
	return server.underlying.Serve(server.Address)
}

// Stop shuts the server down
func (server *Server) Stop() error {
	if server.underlying == nil {
		return util.NewError("already stopped")
	}
	return server.underlying.Shutdown(nil)
}

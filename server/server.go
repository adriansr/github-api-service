package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/adriansr/github-api-service/model"
)

type Server struct {
	address string
	client  model.TopContributorGetter
}

type ApiError struct {
	Error string `json:"error"`
}

const (
	apiPath      = "/search"
	serverName   = "adriansr/github-api-service"
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

func (server *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	setCommonHeaders(writer)

	log.Printf("Serving API")

	// only accept GET requests
	if request.Method != "GET" {
		// required when sending 405 Method Not Allowed
		writer.Header().Add("Allow", "GET")
		sendError(writer, http.StatusMethodNotAllowed, "only GET requests allowed")
		return
	}

	/*if request.URL.Path != apiPath {
		sendError(writer, http.StatusNotFound, "not found")
	}*/
	params := request.URL.Query()
	count, err := strconv.Atoi(params.Get("count"))
	if err != nil {
		count = defaultCount
	}

	if count != 50 && count != 100 && count != 150 {
		sendError(writer, http.StatusBadRequest, "count parameter not valid")
		return
	}

	city := params.Get("city")
	if len(city) == 0 {
		sendError(writer, http.StatusBadRequest, "missing parameter: city")
		return
	}

	result, err := server.client.GetTopContributors(city, count)
	if err != nil {
		sendError(writer, http.StatusInternalServerError,
			"query failed: "+err.Error())
		return
	}
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

func New(address string, client model.TopContributorGetter) (*Server, error) {
	server := &Server{address, client}
	http.Handle(apiPath, server)
	http.HandleFunc("/", notFound)
	log.Printf("Registered API endpoint '%s'", apiPath)
	return server, nil
}

func (server *Server) Run() error {
	log.Printf("Accepting requests at '%s'", server.address)
	return http.ListenAndServe(server.address, nil)
}

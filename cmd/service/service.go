package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/adriansr/github-api-service/config"
	"github.com/adriansr/github-api-service/githubapi"
	"github.com/adriansr/github-api-service/server"
)

const (
	configFilePath = "config.json"
)

func main_Demo1() {
	location := os.Args[1]
	count, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	config, err := config.LoadFile(configFilePath)
	if err != nil {
		log.Fatal(err)
	}
	client, err := githubapi.NewClient(
		config.Credentials.Username,
		config.Credentials.Password,
		config.Client.ApiUrl,
		config.Client.RequestTimeout.Duration)
	if err != nil {
		log.Fatal(err)
	}
	result, err := client.GetTopContributors(location, count)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Got %d results:\n", len(result))
	for idx, elem := range result {
		fmt.Printf("%d: %s (id:%d)\n", idx, elem.Username, elem.ID)
	}
}

func main() {
	config, err := config.LoadFile(configFilePath)
	if err != nil {
		log.Fatal(err)
	}
	client, err := githubapi.NewClient(
		config.Credentials.Username,
		config.Credentials.Password,
		config.Client.ApiUrl,
		config.Client.RequestTimeout.Duration)
	if err != nil {
		log.Fatal("unable to start client: ", err)
	}

	server, err := server.New(config.Server.ListenAddress, client)
	if err != nil {
		log.Fatal("unable to create server: ", err)
	}
	err = server.Run()
	if err != nil {
		log.Fatal("unable to start server: ", err)
	}
}

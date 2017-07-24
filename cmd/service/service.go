package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/adriansr/github-api-service/config"
	"github.com/adriansr/github-api-service/github"
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
	client, err := github.NewClient(
		config.Credentials.Username,
		config.Credentials.Password,
		config.Client.RequestTimeout.Duration)
	if err != nil {
		log.Fatal(err)
	}
	result, err := client.Search(location, count)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Got %d results:\n", len(result))
	for idx, elem := range result {
		fmt.Printf("%d: %s (id:%d)\n", idx, elem.Name, elem.ID)
	}
}

func main() {
	config, err := config.LoadFile(configFilePath)
	if err != nil {
		log.Fatal(err)
	}
	client, err := github.NewClient(
		config.Credentials.Username,
		config.Credentials.Password,
		config.Client.RequestTimeout.Duration)
	if err != nil {
		log.Fatal("unable to start client: ", err)
	}

	log.Printf("Starting server at '%s'", config.Server.ListenAddress)

	server, err := server.New(config.Server.ListenAddress, client)
	if err != nil {
		log.Fatal("unable to create server: ", err)
	}
	err = server.Run()
	if err != nil {
		log.Fatal("unable to start server: ", err)
	}
}

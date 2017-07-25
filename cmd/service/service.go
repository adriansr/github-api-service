// Package main contains the main implementation of the `service` application
package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/adriansr/github-api-service/config"
	"github.com/adriansr/github-api-service/githubapi"
	"github.com/adriansr/github-api-service/server"
)

const (
	configFilePath = "config.json"
)

func main() {
	// load configuration
	config, err := config.LoadFile(configFilePath)
	if err != nil {
		log.Fatal(err)
	}

	// create a client to GitHub API
	client, err := githubapi.NewClient(
		config.Credentials.Username,
		config.Credentials.Password,
		config.Client.ApiUrl,
		config.Client.RequestTimeout.Duration)
	if err != nil {
		log.Fatal("unable to start client: ", err)
	}

	// create our HTTP API server
	server, err := server.New(config.Server.ListenAddress, client)
	if err != nil {
		log.Fatal("unable to create server: ", err)
	}

	// capture SIGINT to support graceful termination with CTRL+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Print("unable to start server: ", err)
			c <- os.Interrupt
		}
	}()

	// wait for termination (signal or server failure)
	<-c

	// terminate
	server.Stop()
	log.Print("Terminated")
}

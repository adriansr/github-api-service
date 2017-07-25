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

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		if err := server.Start(); err != nil {
			log.Print("unable to start server: ", err)
			c <- os.Interrupt
		}
	}()

	<-c

	server.Stop()
	log.Print("Terminated")
}

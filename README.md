# github-api-service
A service to query GitHub top contributors by location using GitHub API and
written in Go

## Table of contents
1. [Prerequisites](#prerequisites)
2. [Download & build](#download)
3. [Running unit tests](#running-unit-tests)
4. [Setup](#setup)
5. [Running the service](#running-the-service)
6. [Performing a query](#performing-a-query)
7. [Stopping the service](#stopping-the-service)
8. [Missing features](#missing-features)
9. [Concurrency and scalability](#concurrency-and-scalability)

## Prerequisites

This program needs Go 1.8 which can be downloaded at https://golang.org/dl/.
It has been tested to work under Linux and macOS.

## Download & build

If you don't already have a `$GOPATH`, just create one:

    $ mkdir go
    $ cd go
    $ export GOPATH=$PWD

At your `$GOPATH` directory, run:

    $ go get github.com/adriansr/github-api-service/cmd/service

The built binary will be at `bin/service`

    $ ./bin/service
    2017/07/26 00:41:54 failed reading configuration file `config.json` [caused by: open config.json: no such file or directory]

## Running unit tests

Use go test to launch tests for all submodules in the project

    $ go test github.com/adriansr/github-api-service/...

## Setup

Copy the provided sample configuration to the current directory:

    $ cp src/github.com/adriansr/github-api-service/sample.config.json config.json

Here you can modify the HTTP server bind address or any other parameter.

    {
        "github_credentials": {
            "username": "",
            "password": ""
        },
        "client": {
            "timeout": "3s",
            "api_url": "https://api.github.com"
        },
        "server": {
            "listen": ":8080"
        }
    }



Associating a GitHub account is optional, but it allows to perform more
queries per second as search limits are pretty low.

## Running the service

With a valid `config.json` the service will now start

    $ ./bin/service
    2017/07/26 00:09:49 Registered API endpoint '/api/top-contributors'
    2017/07/26 00:09:49 Accepting requests at [::]:8080

## Performing a query

The service accepts requests at http://localhost:8080/api/top-contributors.
Two parameters are accepted:

* **city** (mandatory): Name of the city used to filter contributors by the location
advertised in their profile.

* **count** (optional, default 50): Maximum number of top contributors to retrieve.
Valid values are 50, 100 and 150. The latest one is slower as it involves two
requests to GitHub API.

Pass the arguments as GET parameters:

http://localhost:8080/api/top-contributors?city=Barcelona&count=100

Result:

    [{"id":125005,"name":"kristianmandrup"},{"id":663460,"name":"ajsb85"},...]

The output is in JSON format. Consists of a list of objects with an `id` field of integer type (the user's GitHub id) and `name`, a string with the GitHub username.

## Stopping the service

The service can be stopped gracefully by sending it a SIGINT signal. That is
pressing CTRL+C on its running terminal or using `kill -INT`.

## Missing features

Due to limited time available many features have not been implemented:

* Caching of responses per city to avoid querying GitHub repeatedly.

* Correct pagination in the event that GitHub search API lowers its
current maximum of 100 results per query.

* Authentication, as an optional assignment.

## Concurrency and scalability

All Go language features and libraries from the used are concurrent. As Go
already does a great job at handling concurrency with its goroutinges, there
is no need to worry about non-blocking calls to avoid blocking threads.

Currently all goroutines are managed by the http server library, and requests
to GitHub API are performed from the same goroutine that is serving the request.
This means that concurrent requests to this API will perform concurrent requests
to GitHub API, and this might cause problems in the future if GitHub decides to
limit concurrent access. In this case it would be necessary to serialise access
to GitHub API using a fixed worker pool (possibly of just one goroutine). This
pattern is well illustrated in the following blog post:

http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/

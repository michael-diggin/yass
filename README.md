[![Build Status](https://travis-ci.org/michael-diggin/yass.svg?branch=master)](https://travis-ci.org/michael-diggin/yass)

# yass (Yet Another Storage Solution)
A distributed key-value cache service written in golang using gRPC.

This was a simple side project for me to get some understanding of basic distributed systems/gRPC.

The main branch uses an in-memory cache service, the redis-branch uses a redis connection as the key-value store. 

## Structure of the application

This codebase contains both the server/cache side code (contained in /backend), as well as a client side package to connect to the server and cache (contained in /yass). 


## How to Use

This is a go-gettable package so to include in your Go project run:

 `go get github.com/michael-diggin/yass/...` 

From the command line the cache service can be spun up using the command (PORT is a required field)

`yass deploy PORT`

This will build a docker image and run the container exposing the specified port. 

To connect to the cache service and set/get values from it, the client side package exposes a simple API and makes this very easy. 

```golang
package main

import "github.com/michael-diggin/yass/yass"

func main() {

    // accepts a context for dialing the service and the address of service
    cache, _ := yass.NewClient(context.Background(), "localhost:8080")

    defer cache.Close() // terminates the connection

    // Healthcheck endpoint to check that cache is up and running
    ok, err := cache.Ping()
    if !ok {
        ...
    }

    // Place the key/value pair into the cache
    cache.SetValue(context.Background(), "key", "value")

    // Get the value of a key
    val := cache.GetValue(context.Background(), "key") // == "value"

    // Remove the pair from the cache
    cache.DelValue(context.Background(), "key")
}
```

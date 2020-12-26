[![Build Status](https://github.com/michael-diggin/yass/workflows/Build/badge.svg?branch=main)](https://github.com/michael-diggin/yass/actions)

# yass (Yet Another Storage Solution)
A distributed key-value storage service written in golang.

The underlying data structure is a simple golang `map` 

## Structure of the application

This codebase contains both the the storage server code, the api gateway code, as well as the gRPC client code. 


## How to Use (WIP)

This is a go-gettable package so to include in your Go project run:

 `go get github.com/michael-diggin/yass/` 


To connect to the storage service and set/get values from it, the client side package exposes a simple API and makes this very easy. It can also be accessed via a REST endpoint.

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

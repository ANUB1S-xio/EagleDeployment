package main

import (
    "log"
    "github.com/ANUB1s-xio/TidalFlow/framework/api"
)

func main() {
    log.Println("Starting API server on :8080...")
    api.Run() // Initialize the API server
}

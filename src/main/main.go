package main

import (
	"http"
)

func main() {
	server := &http.HttpServer{"/Users/bianwenlong/Downloads", "8080"}
	server.Serve()
}

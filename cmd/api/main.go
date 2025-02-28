package main

import "github.com/gopal-lohar/hackathon-2025/internal/api"

func main() {
	apiServer := api.NewAPIServer()
	// go apiServer.ListenForEndpointConnections() // HACK:
	apiServer.Run()
}

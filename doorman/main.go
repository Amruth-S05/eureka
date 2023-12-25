package main

import (
	"github.com/amruth-s05/eureka/doorman/api"
	"github.com/amruth-s05/eureka/doorman/common"
	"net/http"
)

func main() {
	common.Log("Adding API handlers...")
	http.HandleFunc("/api/feeder", api.FeedHandler)

	common.Log("Starting feeder...")
	api.StartFeederSystem()

	common.Log("Starting Doorman server on port 8080")
	_ = http.ListenAndServe("localhost:8080", nil)
}

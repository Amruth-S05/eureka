package main

import (
	"fmt"
	"github.com/amruth-s05/eureka/doorman/api"
	"github.com/amruth-s05/eureka/doorman/common"
	"net/http"
	"os"
)

func main() {
	common.Log("Adding API handlers...")
	http.HandleFunc("/api/feeder", api.FeedHandler)
	http.HandleFunc("/api/query", api.QueryHandler)

	common.Log("Starting feeder...")
	api.StartFeederSystem()

	port := fmt.Sprintf(":%s", os.Getenv("API_PORT"))
	common.Log(fmt.Sprintf("Starting Doorman server on port %s...\n", port))
	_ = http.ListenAndServe(port, nil)
}

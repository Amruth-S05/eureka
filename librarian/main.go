package main

import (
	"fmt"
	"github.com/amruth-s05/eureka/librarian/api"
	"github.com/amruth-s05/eureka/librarian/common"
	"net/http"
	"os"
)

func main() {
	common.Log("Adding API Handlers...")
	http.HandleFunc("/api/index", api.IndexHandler)
	http.HandleFunc("/api/query", api.QueryHandler)

	common.Log("Starting Eureka Librarian server on port 9090")
	api.StartIndexSystem()

	port := fmt.Sprintf(":%s", os.Getenv("API_PORT"))
	common.Log(fmt.Sprintf("Starting Eureka Librarian server on port %s...\n", port))
	_ = http.ListenAndServe(port, nil)
}

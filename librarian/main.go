package main

import (
	"github.com/amruth-s05/eureka/librarian/api"
	"github.com/amruth-s05/eureka/librarian/common"
	"net/http"
)

func main() {
	common.Log("Adding API Handlers...")
	http.HandleFunc("/api/index", api.IndexHandler)
	http.HandleFunc("/api/query", api.QueryHandler)

	common.Log("Starting Eureka Librarian server on port 9090")
	api.StartIndexSystem()

	common.Log("Starting Eureka Librarian server on port 9090")
	_ = http.ListenAndServe(":9090", nil)
}

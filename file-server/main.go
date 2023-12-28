package main

import (
	"log"
	"net/http"
)

var books = map[string]string{
	"book1": `apple apple cat zebra`,
	"book2": `banana cake zebra`,
	"book3": `apple cake cake whale`,
}

func reqHandler(w http.ResponseWriter, r *http.Request) {
	bookId := r.URL.Path[1:]
	book, _ := books[bookId]
	w.Write([]byte(book))
}

func main() {
	log.Println("Starting File Server on Port 9876...")
	http.HandleFunc("/", reqHandler)
	_ = http.ListenAndServe(":9876", nil)
}

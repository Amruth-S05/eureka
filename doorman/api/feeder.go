package api

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/amruth-s05/eureka/doorman/common"
	"io"
	"net/http"
	"strings"
	"time"
)

type payload struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

type document struct {
	Doc   string `json:"-"`
	Title string `json:"title"`
	DocId string `json:"doc_id"`
}

type token struct {
	Line   string `json:"line"`
	Token  string `json:"token"`
	Title  string `json:"title"`
	DocId  string `json:"doc_id"`
	LIndex int    `json:"l_index"`
	Index  int    `json:"token_index"`
}

type dMsg struct {
	DocId string
	Ch    chan document
}

type lMsg struct {
	LIndex int
	DocId  string
	Ch     chan string
}

type lMeta struct {
	LIndex int
	DocId  string
	Line   string
}

type dAllMsg struct {
	Ch chan []document
}

var (
	// done signals all  listening goroutines to stop
	done chan bool

	// dGetCh is used to retrieve a single document from store
	dGetCh chan dMsg

	// lGetCh is used to retrieve a single line from store
	lGetCh chan lMsg

	// lStoreCh is used to put a line into store
	lStoreCh chan lMeta

	// iAddCh is used to add token to index (Librarian)
	iAddCh chan token

	// dStoreCh is used to put a document into store
	dStoreCh chan document

	// dProcessCh is used to process a document and convert it to tokens
	dProcessCh chan document

	// dGetAllCh is used to retrieve all documents in store
	dGetAllCh chan dAllMsg

	// pProcessCh is used to process the /feeder's payload and start the
	// indexing process
	pProcessCh chan payload
)

// indexAdder adds token to index (Librarian)
func indexAdder(ch chan token, done chan bool) {
	for {
		select {
		case tok := <-ch:
			fmt.Println("adding to librarian:", tok.Token)
		case <-done:
			common.Log("Exiting indexAdder.")
			return
		}
	}
}

// lineStore maintains catalog of all lines for all documents being indexed
func lineStore(ch chan lMeta, callback chan lMsg, done chan bool) {
	store := map[string]string{}
	for {
		select {
		case line := <-ch:
			id := fmt.Sprintf("%s-%d", line.DocId, line.LIndex)
			store[id] = line.Line
		case ch := <-callback:
			line := ""
			id := fmt.Sprintf("%s-%d", ch.DocId, ch.LIndex)
			if l, exists := store[id]; exists {
				line = l
			}
			ch.Ch <- line
		case <-done:
			common.Log("Exiting docStore")
			return
		}
	}
}

// indexProcessor is responsible for converting a document into tokens for indexing
func indexProcessor(ch chan document, lStoreCh chan lMeta, iAddCh chan token, done chan bool) {
	for {
		select {
		case doc := <-ch:
			docLines := strings.Split(doc.Doc, "\n")

			lin := 0
			for _, line := range docLines {
				if strings.TrimSpace(line) == "" {
					continue
				}

				lStoreCh <- lMeta{
					LIndex: lin,
					Line:   line,
					DocId:  doc.DocId,
				}

				index := 0
				words := strings.Fields(line)
				for _, word := range words {
					if tok, valid := common.SimplifyToken(word); valid {
						iAddCh <- token{
							Token:  tok,
							LIndex: lin,
							Line:   line,
							Index:  index,
							DocId:  doc.DocId,
							Title:  doc.Title,
						}
						index++
					}
				}
				lin++
			}
		case <-done:
			common.Log("Exiting indexProcessor.")
			return
		}
	}
}

// docStore maintains a catalog of all documents being indexed.
func docStore(add chan document, get chan dMsg, dGetAllCh chan dAllMsg, done chan bool) {
	store := map[string]document{}

	for {
		select {
		case doc := <-add:
			store[doc.DocId] = doc
		case m := <-get:
			m.Ch <- store[m.DocId]
		case ch := <-dGetAllCh:
			var docs []document
			for _, doc := range store {
				docs = append(docs, doc)
			}
			ch.Ch <- docs
		case <-done:
			common.Log("Exiting docStore.")
			return
		}
	}
}

// docProcessor processes new document payloads
func docProcessor(in chan payload, dStoreCh chan document, dProcessCh chan document, done chan bool) {
	for {
		select {
		case newDoc := <-in:
			var err error
			doc := ""

			if doc, err = getFile(newDoc.URL); err != nil {
				common.Warn(err.Error())
				continue
			}

			titleId := getTitleHash(newDoc.Title)
			msg := document{
				Doc:   doc,
				DocId: titleId,
				Title: newDoc.Title,
			}

			dStoreCh <- msg
			dProcessCh <- msg
		case <-done:
			common.Log("Exiting docProcessor")
			return
		}
	}
}

func getFile(url string) (string, error) {
	var res *http.Response
	var err error

	if res, err = http.Get(url); err != nil {
		errMsg := fmt.Errorf("Unable to retrieve URL: %s.\nError: %s", url, err)
		return "", errMsg
	}
	if res.StatusCode > 200 {
		errMsg := fmt.Errorf("Unable to retrieve URL: %s.\nStatus Code: %d", url, res.StatusCode)
		return "", errMsg
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		errMsg := fmt.Errorf("Error while reading response: URL:%s. \nError: %s", url, err.Error())
		return "", errMsg
	}

	return string(body), nil
}

func getTitleHash(title string) string {
	hash := sha1.New()
	title = strings.ToLower(title)

	str := fmt.Sprintf("%s-%s", time.Now(), title)
	hash.Write([]byte(str))

	hByte := hash.Sum(nil)

	return fmt.Sprintf("%x", hByte)
}

func StartFeederSystem() {
	done = make(chan bool)

	dGetCh = make(chan dMsg, 8)
	dGetAllCh = make(chan dAllMsg)

	iAddCh = make(chan token, 8)
	pProcessCh = make(chan payload, 8)

	dStoreCh = make(chan document, 8)
	dProcessCh = make(chan document, 8)
	lGetCh = make(chan lMsg)
	lStoreCh = make(chan lMeta, 8)

	for i := 0; i < 4; i++ {
		go indexAdder(iAddCh, done)
		go docProcessor(pProcessCh, dStoreCh, dProcessCh, done)
		go indexProcessor(dProcessCh, lStoreCh, iAddCh, done)
	}
	go docStore(dStoreCh, dGetCh, dGetAllCh, done)
	go lineStore(lStoreCh, lGetCh, done)
}

// FeedHandler start processing the payload which contains the file to index
func FeedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		ch := make(chan []document)
		dGetAllCh <- dAllMsg{Ch: ch}
		docs := <-ch
		close(ch)

		if serializedPayload, err := json.Marshal(docs); err == nil {
			w.Write(serializedPayload)
		} else {
			common.Warn("Unable to serialize all docs: " + err.Error())
			w.WriteHeader(500)
			w.Write([]byte(`{"code": 500, "msg": "Error occurred while trying to retrieve documents."}`))
		}
		return
	} else if r.Method != "POST" {
		w.WriteHeader(405)
		w.Write([]byte(`{"code": 405, "msg": "Method Not Allowed."}`))
		return
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var newDoc payload
	err := decoder.Decode(&newDoc)
	if err != nil {
		common.Warn(err.Error())
	}
	pProcessCh <- newDoc

	w.Write([]byte(`{"code": 200, "msg": "Request is being processed."}`))
}

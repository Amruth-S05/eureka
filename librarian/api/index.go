package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// tPayload is used to parse the JSON payload consisting of Token data.
type tPayload struct {
	Token  string `json:"token"`
	Title  string `json:"title"`
	DocId  string `json:"doc_id"`
	LIndex int    `json:"line_index"`
	Index  int    `json:"token_index"`
}

type tIndex struct {
	Index  int
	LIndex int
}

func (ti *tIndex) String() string {
	return fmt.Sprintf("i: %d, li: %d", ti.Index, ti.LIndex)
}

type tIndices []tIndex

type document struct {
	Count   int
	DocId   string
	Title   string
	Indices map[int]tIndices // key represents Line Index
}

func (d *document) String() string {
	str := fmt.Sprintf("%s (%s): %d\n", d.Title, d.DocId, d.Count)
	var buffer bytes.Buffer

	for line, tis := range d.Indices {
		var lBuffer bytes.Buffer
		for _, ti := range tis {
			lBuffer.WriteString(fmt.Sprintf("%s", ti.String()))
		}
		buffer.WriteString(fmt.Sprintf("@%d -> %s\n", line, lBuffer.String()))
	}
	return str + buffer.String()
}

type documentCatalog map[string]*document

func (dc *documentCatalog) String() string {
	return fmt.Sprintf("%#v", dc)
}

type tCatalog map[string]documentCatalog

func (tc *tCatalog) String() string {
	return fmt.Sprintf("%#v", tc)
}

type tcCallback struct {
	Token string
	Ch    chan tcMsg
}

type tcMsg struct {
	Token string
	DC    documentCatalog
}

// pProcessCh is used to process /index's payload and start process
// to add the token to tCatalog
var pProcessCh chan tPayload

// tcGet is used to retrieve a token's catalog (documentCatalog)
var tcGet chan tcCallback

func tIndexer(ch chan tPayload, callback chan tcCallback) {
	store := tCatalog{}
	for {
		select {
		case msg := <-callback:
			dc := store[msg.Token]
			msg.Ch <- tcMsg{
				DC:    dc,
				Token: msg.Token,
			}
		case pd := <-ch:
			dc, exists := store[pd.Token]
			if !exists {
				dc = documentCatalog{}
				store[pd.Token] = dc
			}

			doc, exists := dc[pd.DocId]
			if !exists {
				doc = &document{
					DocId:   pd.DocId,
					Title:   pd.Title,
					Indices: map[int]tIndices{},
				}
				dc[pd.DocId] = doc
			}

			tin := tIndex{
				Index:  pd.Index,
				LIndex: pd.LIndex,
			}
			doc.Indices[tin.LIndex] = append(doc.Indices[tin.LIndex], tin)
			doc.Count++
		}
	}
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		w.WriteHeader(405) // method not allowed
		w.Write([]byte(`{"code": 405, "msg": "Method Not Allowed."}`))
		return
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var tp tPayload
	_ = decoder.Decode(&tp)

	log.Printf("Token recieved%#v\n", tp)

	pProcessCh <- tp

	w.Write([]byte(`{"code": 200, "msg": "Tokens are being added to index"}`))
}

func StartIndexSystem() {
	pProcessCh = make(chan tPayload, 100)
	tcGet = make(chan tcCallback, 100)
	go tIndexer(pProcessCh, tcGet)
}

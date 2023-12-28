package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/amruth-s05/eureka/doorman/common"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
)

var librarianEndpoints = map[string]string{}

func init() {
	librarianEndpoints["lib1"] = os.Getenv("LIB1")
	librarianEndpoints["lib2"] = os.Getenv("LIB2")
	librarianEndpoints["lib3"] = os.Getenv("LIB3")
}

type docs struct {
	DocId string `json:"doc_id"`
	Score int    `json:"doc_score"`
}

type queryResult struct {
	Count int    `json:"count"`
	Data  []docs `json:"data"`
}

func queryLibrarian(endpoint string, stBytes io.Reader, ch chan<- queryResult) {
	resp, err := http.Post(endpoint+"/query", "application/json", stBytes)
	if err != nil {
		common.Warn(fmt.Sprintf("%s -> %+v", endpoint, err))
		ch <- queryResult{}
		return
	}
	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	var qr queryResult
	_ = json.Unmarshal(body, &qr)
	log.Println(fmt.Sprintf("%s -> %#v", endpoint, qr))
	ch <- qr
}

func getResultMap(ch <-chan queryResult) map[string]int {
	var results []docs
	for range librarianEndpoints {
		if result := <-ch; result.Count > 0 {
			results = append(results, result.Data...)
		}
	}

	resultsMap := map[string]int{}
	for _, doc := range results {
		docId := doc.DocId
		score := doc.Score
		if _, exists := resultsMap[docId]; !exists {
			resultsMap[docId] = 0
		}
		resultsMap[docId] += score
	}

	return resultsMap
}

func QueryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405) // method not allowed
		w.Write([]byte(`{"code": 405, ",msg": "Method Not Allowed"}`))
		return
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var searchTerms []string
	if err := decoder.Decode(&searchTerms); err != nil {
		common.Warn("Unable to parse request." + err.Error())

		w.WriteHeader(400) // bad request
		w.Write([]byte(`{"code": 400, "msg": "Unable to parse payload."}`))
		return
	}

	st, err := json.Marshal(searchTerms)
	if err != nil {
		panic(err)
	}
	stBytes := bytes.NewBuffer(st)

	resultsCh := make(chan queryResult)

	for _, le := range librarianEndpoints {
		func(endpoint string) {
			go queryLibrarian(endpoint, stBytes, resultsCh)
		}(le)
	}

	resultsMap := getResultMap(resultsCh)
	close(resultsCh)

	sortedResults := sortResults(resultsMap)

	payload, _ := json.Marshal(sortedResults)
	w.Header().Add("Content-Type", "application/json")
	w.Write(payload)

	fmt.Printf("%#v\n", sortedResults)
}

func sortResults(rm map[string]int) []document {
	scoreMap := map[int][]document{}
	ch := make(chan document)
	for docId, score := range rm {
		if _, exists := scoreMap[score]; !exists {
			scoreMap[score] = []document{}
		}

		dGetCh <- dMsg{
			DocId: docId,
			Ch:    ch,
		}
		doc := <-ch

		scoreMap[score] = append(scoreMap[score], doc)
	}
	close(ch)

	var scores []int
	for score := range scoreMap {
		scores = append(scores, score)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(scores)))

	var sortedResults []document
	for _, score := range scores {
		resDocs := scoreMap[score]
		sortedResults = append(sortedResults, resDocs...)
	}
	return sortedResults
}

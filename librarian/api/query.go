package api

import (
	"encoding/json"
	"github.com/amruth-s05/eureka/librarian/common"
	"net/http"
	"sort"
)

type docResult struct {
	DocId   string   `json:"doc_id"`
	Score   int      `json:"doc_score"`
	Indices tIndices `json:"token_indices"`
}

type result struct {
	Count int         `json:"count"`
	Data  []docResult `json:"data"`
}

// getResults returns unsorted search results & a map of documents
// containing tokens
func getResults(out chan tcMsg, count int) tCatalog {
	tc := tCatalog{}
	for i := 0; i < count; i++ {
		dc := <-out
		tc[dc.Token] = dc.DC
	}
	close(out)

	return tc
}

func getFScores(docIdScore map[string]int) (map[int][]string, []int) {
	// fScore maps frequency score to set of documents
	var fScore map[int][]string

	var fSorted []int

	for dId, score := range docIdScore {
		fs := fScore[score]
		fScore[score] = []string{}
		fScore[score] = append(fs, dId)
		fSorted = append(fSorted, score)
	}

	sort.Sort(sort.Reverse(sort.IntSlice(fSorted)))

	return fScore, fSorted
}

func getDocMaps(tc tCatalog) (map[string]int, map[string]tIndices) {
	//docIdStore maps DocIds to occurrences of all tokens
	// key: DocId
	// val: Sum of all occurrences of tokens do far
	docIdScore := map[string]int{}
	docIndices := map[string]tIndices{}

	// for each token's catalog
	for _, dc := range tc {
		// for each document registered under the token
		for dId, doc := range dc {
			// add to docId score
			var tokIndices tIndices
			for _, tList := range doc.Indices {
				tokIndices = append(tokIndices, tList...)
			}
			docIdScore[dId] += doc.Count
			dti := docIndices[dId]
			docIndices[dId] = append(dti, tokIndices...)
		}
	}
	return docIdScore, docIndices
}

func sortResults(tc tCatalog) []docResult {
	docIdScore, docIndices := getDocMaps(tc)
	fScore, fSorted := getFScores(docIdScore)

	var results []docResult
	addedDocs := map[string]bool{}

	for _, score := range fSorted {
		for _, docId := range fScore[score] {
			if _, exists := addedDocs[docId]; exists {
				continue
			}
			results = append(results, docResult{
				docId, score, docIndices[docId],
			})
			addedDocs[docId] = false
		}
	}
	return results
}

// getSearchResults returns a list of documents
// they are listed in descending order of occurrences
func getSearchResults(sts []string) []docResult {
	callback := make(chan tcMsg)
	for _, st := range sts {
		go func(term string) {
			tcGet <- tcCallback{Token: term, Ch: callback}
		}(st)
	}
	cts := getResults(callback, len(sts))
	results := sortResults(cts)
	return results
}

func QueryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405) // method not allowed
		w.Write([]byte(`{"code": 405}, "msg": "Method Not Allowed."`))
		return
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var searchTerms []string
	_ = decoder.Decode(&searchTerms)

	results := getSearchResults(searchTerms)

	payload := result{
		Count: len(results),
		Data:  results,
	}

	if serializedPayLoad, err := json.Marshal(payload); err == nil {
		w.Header().Add("Content-Type", "application/json")
		w.Write(serializedPayLoad)
	} else {
		common.Warn("Unable to serialize all docs: " + err.Error())
		w.WriteHeader(500) // internal server error
		w.Write([]byte(`{"code": 500, "msg": "Error occurred while trying to retrieve documents."}`))
	}
}

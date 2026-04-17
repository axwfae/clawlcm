package retrieval

import (
	"fmt"
	"sort"
	"strings"

	"github.com/axwfae/clawlcm/store"
	"github.com/axwfae/clawlcm/tokenizer"
)

type BM25Scorer struct {
	k1    float64
	b     float64
	avgdl float64
}

func NewBM25Scorer(avgDocLength float64) *BM25Scorer {
	return &BM25Scorer{
		k1:    1.5,
		b:     0.75,
		avgdl: avgDocLength,
	}
}

func (s *BM25Scorer) Score(query string, docKeywords string) float64 {
	queryTerms := tokenizer.New().Tokenize(query)
	if len(queryTerms) == 0 {
		return 0
	}

	docTerms := strings.Split(docKeywords, ",")
	docFreq := make(map[string]int)
	for _, term := range docTerms {
		if term = strings.TrimSpace(term); term != "" {
			docFreq[term]++
		}
	}

	score := 0.0
	docLen := len(docTerms)
	if docLen == 0 {
		return 0
	}

	for _, qterm := range queryTerms {
		df := docFreq[qterm]
		if df == 0 {
			continue
		}

		termFreq := float64(df)
		numerator := termFreq * (s.k1 + 1)
		denominator := termFreq + s.k1*(1-s.b+s.b*float64(docLen)/s.avgdl)
		score += numerator / denominator
	}

	return score
}

type RetrievalEngine struct {
	scorer *BM25Scorer
	store  *store.Store
}

func NewRetrievalEngine(s *store.Store, avgDocLen float64) *RetrievalEngine {
	return &RetrievalEngine{
		scorer: NewBM25Scorer(avgDocLen),
		store:  s,
	}
}

func (e *RetrievalEngine) Search(conversationID int64, query string, maxResults int, minScore float64) ([]store.SearchResult, error) {
	items, err := e.store.GetContextItems(conversationID)
	if err != nil {
		return nil, fmt.Errorf("get context items for conversation %d: %w", conversationID, err)
	}

	results := make([]store.SearchResult, 0)
	for _, item := range items {
		score := e.scorer.Score(query, item.Keywords)
		if score >= minScore {
			results = append(results, store.SearchResult{
				Item:  item,
				Score: score,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > maxResults {
		results = results[:maxResults]
	}

	return results, nil
}

package tokenizer

import (
	"regexp"
	"strings"
)

type Tokenizer struct {
	stopWords map[string]bool
}

func New() *Tokenizer {
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "being": true, "have": true, "has": true,
		"had": true, "do": true, "does": true, "did": true, "will": true,
		"would": true, "could": true, "should": true, "may": true, "might": true,
		"must": true, "shall": true, "can": true, "need": true, "to": true,
		"of": true, "in": true, "for": true, "on": true, "with": true,
		"at": true, "by": true, "from": true, "as": true, "into": true,
		"through": true, "during": true, "before": true, "after": true,
		"above": true, "below": true, "between": true, "under": true,
		"again": true, "further": true, "then": true, "once": true,
		"here": true, "there": true, "when": true, "where": true, "why": true,
		"how": true, "all": true, "each": true, "few": true, "more": true,
		"most": true, "other": true, "some": true, "such": true, "no": true,
		"nor": true, "not": true, "only": true, "own": true, "same": true,
		"so": true, "than": true, "too": true, "very": true, "just": true,
		"if": true, "else": true, "that": true, "this": true,
		"these": true, "those": true, "what": true, "which": true, "who": true,
		"it": true, "its": true, "they": true, "them": true,
		"their": true, "we": true, "you": true, "your": true, "he": true,
		"she": true, "him": true, "her": true, "his": true, "i": true, "me": true,
		"my": true, "our": true, "us": true, "any": true, "both": true,
		"about": true, "over": true, "out": true, "up": true, "down": true,
		"off": true, "because": true, "until": true,
	}

	return &Tokenizer{stopWords: stopWords}
}

func (t *Tokenizer) Tokenize(text string) []string {
	re := regexp.MustCompile(`[\p{L}\p{M}]+`)
	matches := re.FindAllString(text, -1)

	tokens := make([]string, 0)
	for _, match := range matches {
		lower := strings.ToLower(match)
		if !t.stopWords[lower] && len(match) >= 2 {
			tokens = append(tokens, lower)
		}
	}

	return tokens
}

func (t *Tokenizer) ExtractKeywords(text string, maxKeywords int) string {
	tokens := t.Tokenize(text)

	freq := make(map[string]int)
	for _, token := range tokens {
		freq[token]++
	}

	type kw struct {
		word string
		freq int
	}

	ranked := make([]kw, 0, len(freq))
	for word, count := range freq {
		ranked = append(ranked, kw{word, count})
	}

	for i := 0; i < len(ranked); i++ {
		for j := i + 1; j < len(ranked); j++ {
			if ranked[j].freq > ranked[i].freq {
				ranked[i], ranked[j] = ranked[j], ranked[i]
			}
		}
	}

	result := make([]string, 0)
	for i := 0; i < len(ranked) && i < maxKeywords; i++ {
		result = append(result, ranked[i].word)
	}

	return strings.Join(result, ",")
}

func EstimateTokens(text string) int {
	runeCount := 0
	isCJK := false
	for _, r := range text {
		if r >= 0x4E00 && r <= 0x9FFF {
			isCJK = true
		}
		runeCount++
	}

	if isCJK {
		return runeCount / 2
	}

	avgCharsPerWord := 5
	chars := runeCount
	_ = avgCharsPerWord
	return (chars / avgCharsPerWord) * 4 / 3
}

func EstimateTokensWithConfig(text string, useCJK bool) int {
	if !useCJK {
		words := strings.Fields(text)
		_ = words
		return len(words) * 4 / 3
	}
	return EstimateTokens(text)
}

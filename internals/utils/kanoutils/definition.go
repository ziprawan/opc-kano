package kanoutils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type DictAPITextObject struct {
	Text string `json:"text,omitempty"`
}

type DictAPIInflection struct {
	InflectedForm string `json:"inflectedForm,omitempty"`
}

type DictAPIPronunciation struct {
	AudioFile        string   `json:"audioFile,omitempty"`
	Dialects         []string `json:"dialects,omitempty"`
	PhoneticNotation string   `json:"phoneticNotation,omitempty"`
	PhoneticSpelling string   `json:"phoneticSpelling,omitempty"`
}

type DictAPISense struct {
	Definitions           []string       `json:"definitions,omitempty"`
	CrossReferenceMarkers []string       `json:"crossReferenceMarkers,omitempty"`
	CrossReferences       []string       `json:"crossReference,omitempty"`
	ShortDefinitions      []string       `json:"shortDefinitions,omitempty"`
	SubSenses             []DictAPISense `json:"subSenses,omitempty"`

	Examples        []DictAPITextObject ` json:"examples,omitempty"`
	Synonyms        []DictAPITextObject ` json:"synonyms,omitempty"`
	Registers       []DictAPITextObject ` json:"registers,omitempty"`
	SemanticClasses []DictAPITextObject ` json:"semanticClasses,omitempty"`
	Constructions   []DictAPITextObject ` json:"constructions,omitempty"`
	VariantForms    []DictAPITextObject ` json:"variantForms,omitempty"`
	DomainClasses   []DictAPITextObject ` json:"domainClasses,omitempty"`
}

type DictAPIEntry struct {
	Etymologies    []string               `json:"etymologies,omitempty"`
	Inflections    []DictAPIInflection    `json:"inflections,omitempty"`
	Pronunciations []DictAPIPronunciation `json:"pronunciations,omitempty"`
	Senses         []DictAPISense         `json:"senses,omitempty"`

	Notes []DictAPITextObject `json:"notes,omitempty"`
}

type DictAPILexicalCategory struct {
	Text string `json:"text,omitempty"`
}

type DictAPILexicalEntry struct {
	Derivatives     []DictAPITextObject    `json:"derivatives,omitempty"`
	Entries         []DictAPIEntry         `json:"entries,omitempty"`
	Language        string                 `json:"language,omitempty"`
	LexicalCategory DictAPILexicalCategory `json:"lexicalCategory"`
	Phrases         []DictAPITextObject    `json:"phrases,omitempty"`
}

type DictAPIResult struct {
	Language       string                `json:"language,omitempty"`
	LexicalEntries []DictAPILexicalEntry `json:"lexicalEntries,omitempty"`
	Type           string                `json:"type,omitempty"`
	Word           string                `json:"word,omitempty"`
}

type DictAPIResponse struct {
	Query       string          `json:"query,omitempty"`
	Results     []DictAPIResult `json:"results,omitempty"`
	LastUpdated uint            `json:"last_updated,omitempty"`
}

func FindDefinition(word string) (*DictAPIResponse, error) {
	word = url.PathEscape(word)
	apiUrl := fmt.Sprintf("https://dict-api.com/api/od/%s", word)

	resp, err := http.Get(apiUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned http code %d", resp.StatusCode)
	}

	var dictAPIResponse DictAPIResponse
	err = json.NewDecoder(resp.Body).Decode(&dictAPIResponse)
	if err != nil {
		return nil, err
	}

	return &dictAPIResponse, nil
}

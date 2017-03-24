package search

import (
	"fmt"

	"github.com/spaolacci/murmur3"
)

// Result - Search result.
type Result struct {
	// ID - Search result ID.
	ID string `json:"id,omitempty" gorethink:"id,omitempty"`
	// Query - Search query.
	Query string `json:"query,omitempty" gorethink:"query,omitempty"`
	// Page - Search results page.
	Page int `json:"page,omitempty" gorethink:"page,omitempty"`
	// Results - List of URLs.
	Results []string `json:"results,omitempty" gorethink:"results,omitempty"`
	// Engine - Result engine eq. "google".
	Engine string `json:"engine,omitempty" gorethink:"engine,omitempty"`
	// Source - URL source.
	Source string `json:"source,omitempty" gorethink:"source,omitempty"`
}

// GetID - Gets result ID or creates one.
func GetID(res Result) string {
	if res.ID != "" {
		return res.ID
	}
	return ConstructID(res.Engine, res.Query, res.Page)
}

// ConstructID - Constructs result ID.
func ConstructID(engine, query string, page int) string {
	return fmt.Sprintf("%x", murmur3.Sum64([]byte(fmt.Sprintf("%s:%q:%d", engine, query, page))))
}

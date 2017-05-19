package grawler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Fetcher interface {
	// Fetch returns a reader for the body of the downloaded URL,
	// or error if it could not be downloaded. The caller is
	// responsible for body.Close().
	Fetch(url string) (body io.ReadCloser, err error)
}

// Ensure fetcher is a Fetcher.
var _ Fetcher = (*fetcher)(nil)

type fetcher struct {
	Fetcher
}

// Fetch URL returning the reader to the body of the document, or an
// error if URL could not be fetched. The caller must call Close() on
// the reader to avoid resource leaks.
func (fetcher) Fetch(URL string) (io.ReadCloser, error) {
	resp, err := http.Get(URL)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("fetch failed: HTTP status %d", resp.StatusCode)
		return nil, errors.New(msg)
	}

	return resp.Body, nil
}

// NewHTTPFetcher returns a new Fetcher.
func NewHTTPFetcher() *fetcher {
	return &fetcher{}
}

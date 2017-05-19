package grawler

// FollowLink should return true if URL should be downloaded.
type FollowLink func(URL string) bool

// Page captures the links and assets in an HTML document.
type Page struct {
	// The page that was crawled.
	URL string `json:"URL"`

	// The Fetch.Fetcher() error, if any.
	FetcherError error `json:",FetchError,omitempty"`

	// The set of links.
	Links []string `json:",omitempty"`

	// The set of assets.
	Assets []string `json:",omitempty"`
}

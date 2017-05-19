package grawler_test

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/frobware/grawler"
)

// Number of download workers for grawler.Crawl().
var downloaders = []int{-1, 0, 1, 2, 3, 5, 8, 13, 21, 34}

type PageMap map[string]*grawler.Page

// Match Test<XXX> func name in a package specifier.
//
// For example, it will match TestWithLinksButNoAssets in:
//   github.com/frobware/grawler_test.TestWithLinksButNoAssets
var testNameRegexp = regexp.MustCompile(`\.(Test[\p{L}_\p{N}]+)$`)

// A URL filter to follow URLs that are localhost only.
var localhostOnly = func(URL string) bool {
	link, err := url.Parse(URL)

	if err != nil {
		return false
	}

	return link.Hostname() == "localhost" || link.Hostname() == "127.0.0.1"
}

// Ensure fetcher is a Fetcher.
var _ grawler.Fetcher = (*badFetcher)(nil)

type badFetcher struct {
	grawler.Fetcher

	defaultFetcher grawler.Fetcher
}

func (f badFetcher) Fetch(url string) (io.ReadCloser, error) {
	if strings.Contains(url, "page2.html") {
		return nil, errors.New("page2 went boom")
	}
	return f.defaultFetcher.Fetch(url)
}

func startHTTPServer(t *testing.T, dir string) string {
	rootdir := fmt.Sprintf("testdata/%s", dir)

	if _, err := os.Stat(rootdir); err != nil {
		t.Fatal(err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")

	if err != nil {
		t.Fatal(err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	handler := http.FileServer(http.Dir(rootdir))

	server := &http.Server{
		Addr:    listener.Addr().String(),
		Handler: handler,
	}

	go func() {
		if err := server.Serve(listener); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	return fmt.Sprintf("http://127.0.0.1:%d", port)
}

// Returns the name of the TestXXX function from the call stack.
func testName() string {
	pc := make([]uintptr, 32)
	n := runtime.Callers(0, pc)

	for i := 0; i < n; i++ {
		name := runtime.FuncForPC(pc[i]).Name()
		matches := testNameRegexp.FindStringSubmatch(name)

		if matches == nil {
			continue
		}

		return matches[1]
	}

	panic("test name could not be discovered; increase stack depth?")
}

func testURL(rootURL, pageName string) string {
	return fmt.Sprintf("%s/%s", rootURL, pageName)
}

func crawl(t *testing.T, downloaders int, url string, filter grawler.FollowLink, fetcher grawler.Fetcher) PageMap {
	if fetcher == nil {
		fetcher = grawler.NewHTTPFetcher()
	}

	pages := grawler.Crawl(url, downloaders, fetcher, filter)
	pageMap := make(PageMap, len(pages))

	for _, page := range pages {
		pageMap[page.URL] = page
	}

	return pageMap
}

func assertPagesMissing(t *testing.T, pageMap PageMap, urls []string) {
	for _, url := range urls {
		t.Log("Checking for", url)
		if _, ok := pageMap[url]; ok {
			t.Fatalf("page %q is present", url)
		}
	}
}

func assertPagesFound(t *testing.T, pageMap PageMap, urls []string) {
	for _, url := range urls {
		t.Log("Checking for", url)
		if _, ok := pageMap[url]; !ok {
			t.Fatalf("page %q is missing", url)
		}
	}
}

func assertPageContainsLocalLinks(t *testing.T, rootURL string, page *grawler.Page, names []string) {
	linkSet := make(map[string]bool, len(page.Links))

	for _, link := range page.Links {
		linkSet[link] = true
	}

	for _, name := range names {
		href := fmt.Sprintf("%s/%s", rootURL, name)
		t.Logf("Looking for link %q in page %q", href, page.URL)
		if !linkSet[href] {
			t.Fatalf("link %q not found in page %q", name, page.URL)
		}
	}
}

func assertPageContainsExternalLinks(t *testing.T, rootURL string, page *grawler.Page, names []string) {
	linkSet := make(map[string]bool, len(page.Links))

	for _, link := range page.Links {
		linkSet[link] = true
	}

	for _, name := range names {
		t.Logf("Looking for link %q in page %q", name, page.URL)
		if !linkSet[name] {
			t.Fatalf("link %q not found in page %q", name, page.URL)
		}
	}
}

func assertPageContainsAssets(t *testing.T, rootURL string, page *grawler.Page, names []string) {
	assetSet := make(map[string]bool, len(page.Assets))

	for _, asset := range page.Assets {
		assetSet[asset] = true
	}

	for _, name := range names {
		t.Logf("Looking for asset %q in page %q", name, page.URL)
		if !assetSet[name] {
			t.Fatalf("asset %q not found in page %q, %v", name, page.URL, page)
		}
	}
}

func TestAssetsNoLinks(t *testing.T) {
	rootURL := startHTTPServer(t, testName())

	if testing.Short() {
		downloaders = []int{10}
	}

	for i := range downloaders {
		pageMap := crawl(t, downloaders[i], rootURL, localhostOnly, nil)

		assertPagesFound(t, pageMap, []string{
			rootURL,
			testURL(rootURL, "page1.html"),
		})

		page1 := pageMap[testURL(rootURL, "page1.html")]

		if !reflect.DeepEqual(page1.Links, []string{}) {
			t.Errorf("expected %v, got %v", []string{}, page1.Links)
		}

		expectedAssets := []string{
			"foo.png",
			"/favicon.ico",
			"https://code.jquery.com/jquery-3.2.1.min.js",
		}

		if !reflect.DeepEqual(page1.Assets, expectedAssets) {
			t.Errorf("expected %v, got %v]", expectedAssets, page1.Assets)
		}
	}
}

func TestLinksNoAssets(t *testing.T) {
	rootURL := startHTTPServer(t, testName())

	if testing.Short() {
		downloaders = []int{10}
	}

	for i := range downloaders {
		pageMap := crawl(t, downloaders[i], rootURL, localhostOnly, nil)

		assertPagesFound(t, pageMap, []string{
			rootURL,
			testURL(rootURL, "page1.html"),
			testURL(rootURL, "page2.html"),
		})

		page := pageMap[testURL(rootURL, "page1.html")]

		expectedLinks := []string{
			testURL(rootURL, "page2.html"),
			testURL(rootURL, "page3.html"),
			testURL(rootURL, "page4.html"),
		}

		if !reflect.DeepEqual(page.Links, expectedLinks) {
			t.Errorf("expected %v, got %v", expectedLinks, page.Links)
		}

		if !reflect.DeepEqual(page.Assets, []string{}) {
			t.Errorf("expected %v, got %v", []string{}, page.Assets)
		}

		// page2 is not present, so it should have no links.
		page = pageMap[testURL(rootURL, "page2.html")]

		if !reflect.DeepEqual(page.Links, []string{}) {
			t.Errorf("expected %v, got %v", []string{}, page.Links)
		}

		if !reflect.DeepEqual(page.Assets, []string{}) {
			t.Errorf("expected %v, got %v", []string{}, page.Assets)
		}
	}
}

func TestLinksAndLoops(t *testing.T) {
	rootURL := startHTTPServer(t, testName())

	if testing.Short() {
		downloaders = []int{10}
	}

	for i := range downloaders {
		pageMap := crawl(t, downloaders[i], rootURL, localhostOnly, nil)

		assertPagesFound(t, pageMap, []string{
			rootURL,
			testURL(rootURL, "page1.html"),
			testURL(rootURL, "page2.html"),
		})
	}
}

func TestMultipleLinks(t *testing.T) {
	rootURL := startHTTPServer(t, testName())

	if testing.Short() {
		downloaders = []int{10}
	}

	for i := range downloaders {
		pageMap := crawl(t, downloaders[i], rootURL, localhostOnly, nil)

		assertPagesFound(t, pageMap, []string{
			rootURL,
			testURL(rootURL, "page1.html"),
			testURL(rootURL, "page2.html"),
			testURL(rootURL, "page3.html"),
			testURL(rootURL, "page4.html"),
			testURL(rootURL, "page5.html"),
			testURL(rootURL, "page6.html"),
		})

		page1 := pageMap[testURL(rootURL, "page1.html")]
		page2 := pageMap[testURL(rootURL, "page2.html")]
		page3 := pageMap[testURL(rootURL, "page3.html")]
		page4 := pageMap[testURL(rootURL, "page4.html")]
		page5 := pageMap[testURL(rootURL, "page5.html")]
		page6 := pageMap[testURL(rootURL, "page6.html")]

		assertPageContainsLocalLinks(t, rootURL, page1, []string{"page2.html"})
		assertPageContainsLocalLinks(t, rootURL, page2, []string{"page3.html"})
		assertPageContainsLocalLinks(t, rootURL, page3, []string{"page4.html"})
		assertPageContainsLocalLinks(t, rootURL, page4, []string{"page5.html"})
		assertPageContainsLocalLinks(t, rootURL, page5, []string{
			"page1.html",
			"page2.html",
			"page3.html",
			"page4.html",
			"page5.html",
			"page6.html",
		})

		apple := "http://apple.com"
		ebay := "http://ebay.com"
		facebook := "http://facebook.com"
		google := "http://google.com"
		monzo := "http://monzo.com"

		assertPageContainsExternalLinks(t, rootURL, page1, []string{google})
		assertPageContainsExternalLinks(t, rootURL, page2, []string{ebay})
		assertPageContainsExternalLinks(t, rootURL, page3, []string{facebook})
		assertPageContainsExternalLinks(t, rootURL, page4, []string{apple})
		assertPageContainsExternalLinks(t, rootURL, page5, []string{
			google,
			ebay,
			facebook,
			apple,
			monzo})

		if page6.FetcherError == nil {
			t.Fatal("page6.html should have an error")
		}

		expectedError := errors.New("fetch failed: HTTP status 404")

		if page6.FetcherError.Error() != expectedError.Error() {
			t.Errorf("expected %q, got %q", expectedError, page6.FetcherError)
		}
	}
}

func TestMultipleAssets(t *testing.T) {
	rootURL := startHTTPServer(t, testName())

	if testing.Short() {
		downloaders = []int{10}
	}

	for i := range downloaders {
		pageMap := crawl(t, downloaders[i], rootURL, localhostOnly, nil)

		assertPagesFound(t, pageMap, []string{
			rootURL,
			testURL(rootURL, "page1.html"),
			testURL(rootURL, "page2.html"),
		})

		page1 := pageMap[testURL(rootURL, "page1.html")]
		page2 := pageMap[testURL(rootURL, "page2.html")]

		assertPageContainsAssets(t, rootURL, page1, []string{"image1.png"})
		assertPageContainsAssets(t, rootURL, page2, []string{
			"image2.png",
			"image3.png",
			"image4.png",
		})
	}
}

func TestNoLinksNoAssets(t *testing.T) {
	rootURL := startHTTPServer(t, testName())

	if testing.Short() {
		downloaders = []int{10}
	}

	for i := range downloaders {
		pageMap := crawl(t, downloaders[i], rootURL, localhostOnly, nil)
		gitKeep := testURL(rootURL, ".gitkeep")

		// The site should be empty but we need a token file
		// to keep Git happy. We frob reality and remove it
		// explicitly.
		delete(pageMap, gitKeep)

		if len(pageMap) != 1 {
			t.Errorf("found %d pages, expected 1", len(pageMap))
		}

		assertPagesFound(t, pageMap, []string{
			rootURL,
		})

		index := pageMap[rootURL]

		expectedLinks := []string{
			gitKeep, // necessary evil, should be 0 links.
		}

		if !reflect.DeepEqual(index.Links, expectedLinks) {
			t.Errorf("page links [%v] do not match expected links [%v]", index.Links, expectedLinks)
		}

		expectedAssets := []string{}

		if !reflect.DeepEqual(index.Assets, expectedAssets) {
			t.Errorf("page assets [%v] do not match expected assets [%v]", index.Assets, expectedAssets)
		}
	}
}

func TestFetchError(t *testing.T) {
	rootURL := startHTTPServer(t, testName())

	if testing.Short() {
		downloaders = []int{10}
	}

	for i := range downloaders {
		pageMap := crawl(t, downloaders[i], rootURL, localhostOnly, badFetcher{
			defaultFetcher: grawler.NewHTTPFetcher(),
		})

		assertPagesFound(t, pageMap, []string{
			rootURL,
			testURL(rootURL, "page1.html"),
			testURL(rootURL, "page2.html"),
		})

		page1 := pageMap[testURL(rootURL, "page1.html")]
		page2 := pageMap[testURL(rootURL, "page2.html")]

		if page1.FetcherError != nil {
			t.Errorf("page1.html should not have an error")
		}

		if page2.FetcherError == nil {
			t.Errorf("page2.html should have an error")
		}

		expectedError := errors.New("page2 went boom")

		if page2.FetcherError.Error() != expectedError.Error() {
			t.Errorf("expected %q, got %q", expectedError, page2.FetcherError)
		}
	}
}

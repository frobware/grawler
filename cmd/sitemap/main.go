package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/frobware/grawler"
)

var concurrency int

func init() {
	flag.IntVar(&concurrency, "j", 50, "number of concurrent downloads")
}

var usage = func() {
	fmt.Fprintf(os.Stderr, "%s [ -j <concurrency> ] URL\n", os.Args[0])
	flag.PrintDefaults()
}

func sameHost(base *url.URL, URL string) bool {
	u, err := url.Parse(URL)

	if err != nil {
		return false
	}

	return base.Hostname() == u.Hostname()
}

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 {
		usage()
		os.Exit(1)
	}

	baseURL, err := url.Parse(flag.Arg(0))

	if err != nil {
		log.Fatal(err)
	}

	filter := func(URL string) bool {
		return sameHost(baseURL, URL)
	}

	fetcher := grawler.NewHTTPFetcher()
	pages := grawler.Crawl(flag.Arg(0), concurrency, fetcher, filter)
	b, err := json.MarshalIndent(pages, "", "  ")

	if err != nil {
		log.Fatal("error:", err)
	}

	fmt.Println(string(b))
}

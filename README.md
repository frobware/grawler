[![Travis CI](https://travis-ci.org/frobware/grawler.svg?branch=master)](https://travis-ci.org/frobware/grawler)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/frobware/grawler)
[![Coverage Status](http://codecov.io/github/frobware/grawler/coverage.svg?branch=master)](http://codecov.io/github/frobware/grawler?branch=master)
[![Report Card](https://goreportcard.com/badge/github.com/frobware/grawler)](https://goreportcard.com/report/github.com/frobware/grawler)

# Web Crawler

An example of webcrawling using Go.

## Installation

	$ go get -u golang.org/x/net/html
    $ go get -u github.com/frobware/grawler/...

The binary `sitemap` is an example of using the library. Given a URL
it will print a basic sitemap for the given domain. And it will only
download links within the domain. For example:

    $ sitemap http://gopl.io
	
By default it will use 50 workers to download links concurrently. You
can change that number using the `-j <N>` argument. For example:

    $ sitemap -j 1 http://gopl.io

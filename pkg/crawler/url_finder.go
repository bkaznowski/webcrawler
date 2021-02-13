package crawler

import (
	"log"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

type urlFinder interface {
	find(*url.URL) ([]*url.URL, error)
}

type urlFinderImpl struct {}

func (*urlFinderImpl) find(site *url.URL) ([]*url.URL, error) {
	response, err := http.Get(site.String())
	if err != nil {
		return nil, errors.Wrapf(err, "could not get site \"%v\"", site)
	}
	defer response.Body.Close()

	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "error loading HTTP response body for site \"%v\"", site)
	}

	urls := []*url.URL{}
	document.Find("a").Each(
		func(index int, element *goquery.Selection) {
			href, exists := element.Attr("href")
			if exists {
				hrefURL, err := url.Parse(href)
				if err != nil {
					log.Printf("link not parsable \"%v\". Skipping", href)
					return
				}
				urls = append(urls, hrefURL)
			}
		},
	)
	return urls, nil
}

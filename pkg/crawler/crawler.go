package crawler

import (
	"log"
	"net/url"

	"github.com/bkaznowski/webcrawler/pkg/queue"
	"github.com/pkg/errors"
)

type crawler struct {
	maxWorkers       int
	availableWorkers int
	visited          map[string]bool
	includeExternal  bool
	workerInCh       chan workerInput
	workerOutCh      chan workerOutput
	urlFinder        urlFinder
}

type Crawler interface {
	Crawl(string) ([]Result, error)
}

type Result struct {
	Parent   *url.URL
	Children []*url.URL
}

func New(workers int) Crawler {
	return newCrawler(workers, &urlFinderImpl{})
}

func newCrawler(workers int, finder urlFinder) Crawler {
	workerInCh := make(chan workerInput)
	workerOutCh := make(chan workerOutput)
	for i := 0; i < workers; i++ {
		w := worker{
			workerInputCh:  workerInCh,
			workerOutputCh: workerOutCh,
			urlFinder:      finder,
		}
		go w.work()
	}
	return &crawler{
		maxWorkers:       workers,
		availableWorkers: workers,
		visited:          map[string]bool{},
		workerInCh:       workerInCh,
		workerOutCh:      workerOutCh,
		urlFinder:        finder,
	}
}

func urlsToInterfaces(urls []*url.URL) []interface{} {
	converted := []interface{}{}
	for _, url := range urls {
		converted = append(converted, url)
	}
	return converted
}

func (c *crawler) Crawl(target string) ([]Result, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse \"%v\"", target)
	}
	// Assume the schema is correct and that the domain is now valid. Also assume it is just a domain.

	allResults := []Result{}
	q := queue.New()
	q.Push(targetURL)
	c.visited[cleanUrl(*targetURL)] = true
	for site, workAvailable := q.Pop(); workAvailable || c.availableWorkers != c.maxWorkers; site, workAvailable = q.Pop() {
		if workAvailable {
			if c.availableWorkers == 0 {
				results, visitableChildren, err := c.waitForWorker(targetURL)
				if err != nil {
					return nil, err
				}
				allResults = append(allResults, results)
				q.Push(urlsToInterfaces(visitableChildren)...)
			}
			siteURL := site.(*url.URL)
			c.process(siteURL)
		} else {
			results, visitableChildren, err := c.waitForWorker(targetURL)
			if err != nil {
				return nil, err
			}
			allResults = append(allResults, results)
			q.Push(urlsToInterfaces(visitableChildren)...)
		}
	}
	return allResults, nil
}

func (c *crawler) waitForWorker(target *url.URL) (Result, []*url.URL, error) {
	workerOut := <-c.workerOutCh
	c.availableWorkers++
	if workerOut.err != nil {
		return Result{}, nil, workerOut.err
	}
	workerOut.result = applyDomainIfMissing(workerOut.result, *target)
	result := Result{
		Parent:   workerOut.result.Parent,
		Children: []*url.URL{},
	}
	visitableChildren := []*url.URL{}
	for _, child := range workerOut.result.Children {
		isVisited, isTargetable, err := c.shouldVisit(child, target)
		if err != nil {
			return Result{}, nil, err
		}
		if isTargetable {
			result.Children = append(result.Children, child)
			if !isVisited {
				visitableChildren = append(visitableChildren, child)
			}
		}
	}
	return result, visitableChildren, nil
}

func applyDomainIfMissing(result Result, target url.URL) Result {
	for i, child := range result.Children {
		// should never fail as target is already validated
		appliedDomain, _ := target.Parse(child.String())
		result.Children[i] = appliedDomain
	}
	return result
}

func (c *crawler) process(parent *url.URL) {
	c.availableWorkers--
	c.workerInCh <- workerInput{
		parent: parent,
	}
}

func (c *crawler) shouldVisit(site, target *url.URL) (bool, bool, error) {
	cleanedURL := cleanUrl(*site)
	_, isVisited := c.visited[cleanedURL]
	isTargetable, err := c.isTargetable(site, target)
	if !isVisited && isTargetable {
		c.visited[cleanedURL] = true
		if len(c.visited)%100 == 0 {
			log.Printf("Found %v unique sites already... last one found is %v", len(c.visited), cleanedURL)
		}
	}
	return isVisited, isTargetable, err
}

func (c *crawler) isTargetable(site, target *url.URL) (bool, error) {
	if !(site.Scheme != "http") && !(site.Scheme != "https") {
		return false, nil
	}

	return site.Host == target.Host, nil
}

func cleanUrl(site url.URL) string {
	site.RawQuery = ""
	site.Fragment = ""
	return site.String()
}

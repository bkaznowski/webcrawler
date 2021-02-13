package main

import (
	"fmt"
	"log"
	"flag"

	"github.com/bkaznowski/webcrawler/pkg/crawler"
)

func main() {
	target := flag.String("target", "", "the target website")
	workers := flag.Int("workers", 4, "the number of workers")
	flag.Parse()

	c := crawler.New(*workers)
	results, err := c.Crawl(*target)
	if err != nil {
		log.Fatalf("Something went wrong: %v", err)
	}
	for _, r := range results {
		fmt.Printf("%v %v\n", r.Parent, r.Children)
	}
}
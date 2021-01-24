package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/MetalBlueberry/go-minions/pkg/minions"
	"github.com/PuerkitoBio/goquery"
)

func main() {
	lord := minions.NewLord()
	// There is a lock condition that happens when this queue is full, but anyway...
	quest := make(chan minions.Worker, 1000)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	lord.StartQuest(
		ctx,
		4,
		quest,
	)

	found := make(chan DomainCrawler)
	links := map[string]int{}
	go func() {
		for {
			crawl := <-found
			if _, visited := links[crawl.URL]; !visited {
				log.Printf("visited %s", crawl.URL)
				quest <- crawl
			} else {
				log.Printf("skip, already visited %s", crawl.URL)
			}
			links[crawl.URL]++
		}
	}()

	crawler := DomainCrawler{
		Client: http.DefaultClient,
		Domain: "github.com",
		Depth:  -1,
		Found:  found,
		URL:    "https://github.com/MetalBlueberry",
	}
	found <- crawler

	lord.Wait()

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(links)
}

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type DomainCrawler struct {
	Client Doer
	Domain string
	URL    string
	Depth  int
	Found  chan<- DomainCrawler
}

func (crawler DomainCrawler) Navigate(target string) DomainCrawler {
	return DomainCrawler{
		URL:   target,
		Depth: crawler.Depth - 1,

		Client: crawler.Client,
		Domain: crawler.Domain,
		Found:  crawler.Found,
	}

}

func (crawler DomainCrawler) Work(ctx context.Context) {
	log.Printf("Crawling %s", crawler.URL)
	parsed, err := url.Parse(crawler.URL)
	if err != nil {
		log.Printf("Cannot parse url %s, %s", crawler.URL, err)
		return
	}

	if parsed.Host != crawler.Domain {
		log.Printf("invalid domain %s", crawler.URL)
		return
	}

	crawler.Depth--
	if crawler.Depth == 0 {
		log.Printf("Depth reached %s", crawler.URL)
		return
	}

	request, err := http.NewRequest("GET", crawler.URL, nil)
	if err != nil {
		log.Printf("Cannot create request url %s, %s", crawler.URL, err)
		return
	}

	response, err := crawler.Client.Do(request.WithContext(ctx))
	if err != nil {
		log.Printf("Cannot DO GET url %s, %s", crawler.URL, err)
		return
	}
	defer response.Body.Close()

	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Printf("Cannot parse request body %s, %s", crawler.URL, err)
		return
	}

	document.Find("a").EachWithBreak(func(i int, s *goquery.Selection) bool {
		res, _ := s.Attr("href")
		log.Printf("found %s", res)
		found, err := url.Parse(res)
		if err != nil {
			return true
		}
		if found.Host == "" {
			found.Host = parsed.Host
		}
		if found.Scheme == "" {
			found.Scheme = parsed.Scheme
		}

		select {
		case crawler.Found <- crawler.Navigate(found.String()):
			return true
		case <-ctx.Done():
			return false
		}
	})
	log.Printf("done %s", crawler.URL)

}

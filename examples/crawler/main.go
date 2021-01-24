package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/MetalBlueberry/go-minions/pkg/minions"
	"github.com/PuerkitoBio/goquery"
)

func main() {
	lord := minions.NewLord()
	// There is a lock condition that happens when this queue is full, but anyway...
	quest := make(chan minions.Worker, 1000)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	lord.StartQuest(
		ctx,
		4,
		quest,
	)

	wg := &sync.WaitGroup{}

	found := make(chan DomainCrawler)
	links := map[string]int{}
	go func() {
		for {
			crawl, open := <-found
			if !open {
				log.Print("no more found, finishing quest")
				close(quest)
				return
			}
			if _, visited := links[crawl.URL]; !visited {
				log.Printf("visiting %s", crawl.URL)
				wg.Add(1)
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
		Depth:  2,
		Found:  found,
		URL:    "https://github.com/MetalBlueberry",
		wg:     wg,
	}

	wg.Add(1)
	quest <- crawler
	go func() {
		wg.Wait()
		close(found)
	}()
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
	wg     *sync.WaitGroup
}

func (crawler DomainCrawler) Navigate(target string) DomainCrawler {
	return DomainCrawler{
		URL:   target,
		Depth: crawler.Depth - 1,

		Client: crawler.Client,
		Domain: crawler.Domain,
		Found:  crawler.Found,
		wg:     crawler.wg,
	}

}

func (crawler DomainCrawler) Work(ctx context.Context) {
	log.Printf("Crawling %s", crawler.URL)

	defer crawler.wg.Done()

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

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

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

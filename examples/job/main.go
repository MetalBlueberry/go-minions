package main

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/MetalBlueberry/go-minions/pkg/minions"
	"github.com/MetalBlueberry/go-minions/pkg/quests"
)

type Status int

const (
	// Pending state
	Pending Status = iota
	// Running state
	Running
	// Done state
	Done
	// // Cancelled state
	// Cancelled
)

type Job struct {
	ID       string           `json:"id,omitempty"`
	Status   Status           `json:"status,omitempty"`
	Tags     []string         `json:"tags,omitempty"`
	Start    time.Time        `json:"start,omitempty"`
	End      time.Time        `json:"end,omitempty"`
	finisher minions.Finisher `json:"-"`
	Job      minions.Worker   `json:"-"`
}

func (job *Job) Work(ctx context.Context) {
	// notify individual job is finished
	defer job.finisher.Finish()

	// Track start & end time
	job.Start = time.Now()
	defer func() { job.End = time.Now() }()

	job.Status = Running

	// Run job
	job.Job.Work(ctx)

	job.Status = Done
}

func (job *Job) Wait(ctx context.Context) error {
	return job.finisher.Wait(ctx)
}

func (job *Job) String() string {
	data, _ := json.MarshalIndent(job, "", "  ")
	return string(data)
}

func downloadLinux() {
	go11517, err := os.Create("go1.15.7.linux-amd64.tar.gz")
	if err != nil {
		panic(err)
	}
	defer go11517.Close()

	downloadJob := quests.NewFileDownload(
		"https://golang.org/dl/go1.15.7.linux-amd64.tar.gz",
		go11517,
	)

	timedJob := &Job{
		ID:   "download go1.15.7.linux-amd64.tar.gz",
		Tags: []string{"downloads"},
		Job:  downloadJob,
	}
	lord := minions.NewLord()

	log.Print("sending minions to download golang")
	lord.StartQuest(context.Background(), 1, minions.NewQuest([]minions.Worker{
		timedJob,
	}))

	err = downloadJob.Wait(context.Background())
	if err != nil {
		panic(err)
	}

	if downloadJob.Err != nil {
		log.Printf("Downlad failed, %s, %s", downloadJob.Err, timedJob)
		return
	}

	log.Printf("Downlad finished successfully, %s", timedJob)
}

func NewDownloadJob(URL string) (*Job, error) {
	parsed, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}
	go11517, err := os.Create(filepath.Base(parsed.Path))
	if err != nil {
		panic(err)
	}

	downloadJob := quests.NewFileDownload(
		URL,
		go11517,
	)

	go func() {
		downloadJob.Wait(context.Background())
		go11517.Close()
	}()

	job := &Job{
		ID:   URL,
		Tags: []string{"downloads"},
		Job:  downloadJob,
	}
	return job, nil
}

type Jobs []*Job

func (jobs Jobs) String() string {
	type state interface {
		Status() error
	}
	type status struct {
		ID     string `json:"id,omitempty"`
		Status Status `json:"status"`
		Err    string `json:"err,omitempty"`
	}

	states := make([]status, len(jobs))
	for i := range jobs {

		states[i].ID = jobs[i].ID
		states[i].Status = jobs[i].Status

		if s, ok := jobs[i].Job.(state); ok {
			err := s.Status()
			if err != nil {
				states[i].Err = err.Error()
			}
		}
	}

	data, _ := json.MarshalIndent(states, "", "  ")
	return string(data)
}

func (jobs Jobs) Quest() <-chan minions.Worker {
	buf := make(chan minions.Worker, len(jobs))
	for i := range jobs {
		buf <- jobs[i]
	}
	close(buf)
	return buf
}

func downloadAll() {
	links := []string{
		"https://golang.org/dl/go1.15.7.linux-amd64.tar.gz",
		"https://golang.org/dl/go1.15.7.windows-amd64.msi",
		"https://golang.org/dl/invalid",
		"https://golang.org/dl/go1.15.7.darwin-amd64.pkg",
		"https://golang.org/dl/go1.15.7.src.tar.gz",
	}

	jobs := make(Jobs, len(links))
	for i := range links {
		q, err := NewDownloadJob(links[i])
		if err != nil {
			panic(err)
		}
		jobs[i] = q
	}

	lord := minions.NewLord()

	log.Print("sending minions to download golang")
	lord.StartQuest(context.Background(), 2, jobs.Quest())

	done := make(chan struct{})
	go func() {
		lord.Wait()
		close(done)
	}()

	wait(done, jobs)

	log.Printf("Downlad finished, %s", jobs)
}

func wait(done chan struct{}, jobs Jobs) {
	for {
		select {
		case <-done:
			return
		case <-time.Tick(time.Second):
			log.Printf("Downloading, %s", jobs)
		}
	}
}

func main() {
	downloadLinux()
	downloadAll()
}

package quests

import (
	"context"
	"io"
	"net/http"

	"github.com/MetalBlueberry/go-minions/pkg/minions"
)

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// FileDownload downloads a file from the given url to a io.WriteCloser
type FileDownload struct {
	URL     string
	Storage io.WriteCloser
	Client  Doer
	Err     error

	finisher minions.Finisher
}

// NewFileDownload creates a file download with the default client
func NewFileDownload(url string, storage io.WriteCloser) *FileDownload {
	return &FileDownload{
		URL:     url,
		Storage: storage,
		Client:  http.DefaultClient,
	}
}

// Work implements minions.Worker interface
func (worker *FileDownload) Work(ctx context.Context) {
	defer worker.finisher.Finish()

	req, err := http.NewRequest("GET", worker.URL, nil)
	if err != nil {
		worker.Err = err
		return
	}
	res, err := worker.Client.Do(req.WithContext(ctx))
	if err != nil {
		worker.Err = err
		return
	}
	defer res.Body.Close()
	defer worker.Storage.Close()

	_, err = io.Copy(worker.Storage, res.Body)
	if err != nil {
		worker.Err = err
		return
	}
}

// Wait blocks until the download is finished
func (worker *FileDownload) Wait(ctx context.Context) error {
	return worker.finisher.Wait(ctx)
}

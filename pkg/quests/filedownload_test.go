package quests

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/MetalBlueberry/go-minions/pkg/minions"
	"github.com/stretchr/testify/assert"
)

type MockDoer struct {
	ExpectURL string
	Response  io.Reader
	T         *testing.T
}

func (doer MockDoer) Do(req *http.Request) (*http.Response, error) {
	assert.Equal(doer.T, doer.ExpectURL, req.URL.String())
	return &http.Response{
		Body: ioutil.NopCloser(doer.Response),
	}, nil
}

type BufCloser struct {
	bytes.Buffer
}

func (BufCloser) Close() error { return nil }

func TestFileDownload(t *testing.T) {
	storage := &BufCloser{}

	work := &FileDownload{
		URL:     "www.test.com",
		Storage: storage,
		Client: MockDoer{
			ExpectURL: "www.test.com",
			Response:  strings.NewReader("file content"),
			T:         t,
		},
	}
	works := []minions.Worker{
		work,
	}

	lord := minions.NewLord()
	lord.StartQuest(context.Background(), 1, minions.NewQuest(works))

	err := work.Wait(context.Background())
	assert.Nil(t, err)

	lord.Wait()

	assert.Equal(t, storage.String(), "file content")
}

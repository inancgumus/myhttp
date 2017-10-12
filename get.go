package myhttp

import (
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const (
	defaultTimeout = time.Second * 3
)

// HTTPGetter does a get request to a http endpoint
type HTTPGetter interface {
	Get(url string) (*http.Response, error)
	WrapGet(url string, do func(r *http.Response) error) error
}

// Getter are for implementing HTTPGetter interface
// and reserved for the future work
type Getter struct {
	timeout time.Duration
	client  *http.Client
}

// New creates a new Getter
func New(to time.Duration) *Getter {
	if to == 0 {
		to = defaultTimeout
	}

	// http.Client.Get reuses the transport. this should be created once.
	tp := http.Transport{}

	tp.DialContext = (&net.Dialer{
		Timeout: to,
	}).DialContext

	tp.TLSHandshakeTimeout = to
	tp.ResponseHeaderTimeout = to
	tp.ExpectContinueTimeout = to

	// this should be used everytime. no need to create new one for each Get.
	c := http.Client{
		Transport: &tp,
	}

	return &Getter{client: &c}
}

// Get fetches url with a timeout
func (g *Getter) Get(url string) (*http.Response, error) {
	return g.client.Get(url)
}

// WrapGet gets from `url` and closes the body automatically after running in `do`
func (g *Getter) WrapGet(url string, do func(r *http.Response) error) error {
	r, err := g.Get(url)
	if err != nil {
		return errors.WithStack(err)
	}
	defer r.Body.Close()
	return do(r)
}

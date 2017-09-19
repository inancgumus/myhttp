package myhttp_test

import (
	"errors"
	"io"
	"io/ioutil"
	"myhttp"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
)

// 0 ==> use the default timeout of myhttp
const timeout = 0

func TestHttpGet(t *testing.T) {
	t.Parallel()

	t.Run("good response", func(t *testing.T) {
		h := &handlers{}
		s := newServer(h.code200)
		defer s.Close()

		res, err := myhttp.New(timeout).Get(s.URL)

		assert.NoError(t, err)
		assert.True(t, h.called)
		assert.Equal(t, code200Response, bodyClose(res))
	})

	t.Run("delaying response", func(t *testing.T) {
		s := newServer((&handlers{timeout: time.Millisecond * 5}).delay)
		defer s.Close()

		_, err := myhttp.New(time.Millisecond).Get(s.URL)

		assert.Error(t, err)
		assert.Contains(t, strings.ToLower(err.Error()), "timeout")
	})

	t.Run("bad url", func(t *testing.T) {
		_, err := myhttp.New(timeout).Get("http://wrong")
		assert.Error(t, err)
	})

	t.Run("no data received", func(t *testing.T) {
		t.SkipNow()
	})
}

func TestWrapGet(t *testing.T) {
	t.Run("good response", func(t *testing.T) {
		h := &handlers{}
		s := newServer(h.code200)
		defer s.Close()

		err := myhttp.New(timeout).WrapGet(s.URL, func(r *http.Response) error {
			assert.Equal(t, code200Response, body(r))
			return nil
		})

		assert.NoError(t, err)
		assert.True(t, h.called)
	})

	t.Run("delaying response", func(t *testing.T) {
		s := newServer((&handlers{timeout: time.Millisecond * 5}).delay)
		defer s.Close()

		err := myhttp.New(time.Millisecond).
			WrapGet(s.URL, func(r *http.Response) error {
				return nil
			})

		assert.Error(t, err)
	})

	t.Run("bad do func", func(t *testing.T) {
		s := new200Server()
		defer s.Close()

		rerr := errors.New("returned err")
		err := myhttp.New(timeout).WrapGet(s.URL, func(r *http.Response) error {
			return rerr
		})

		assert.Error(t, err)
		assert.EqualError(t, err, rerr.Error())
	})

	t.Run("closes resp body", func(t *testing.T) {
		s := new200Server()
		defer s.Close()

		var res *http.Response
		err := myhttp.New(timeout).WrapGet(s.URL, func(r *http.Response) error {
			res = r
			return nil
		})

		// check that the body is automatically closed after `do`
		buf := make([]byte, 1024)
		n, err := res.Body.Read(buf)
		assert.EqualError(t, err, "http: read on closed response body")
		assert.Equal(t, 0, n)
	})
}

func TestLeakingConnections(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	mht := myhttp.New(time.Second * 10)

	s := new200Server()
	defer s.Close()

	cur := tcps(t)
	for i := 0; i < 10; i++ {
		res, _ := mht.Get(s.URL)
		// this is the ultimate necessary. without reading the body
		// `http` is dropping the connections to protect its clents!
		bodyClose(res)
	}

	for tries := 10; tries >= 0; tries-- {
		growth := tcps(t) - cur
		if growth > 5 {
			t.Error("leaked")
			return
		}
	}
}

// find tcp connections
func tcps(t *testing.T) (conns int) {
	lsof, err := exec.Command("lsof", "-n", "-p", strconv.Itoa(os.Getpid())).Output()
	if err != nil {
		t.Skip("skipping test; error finding or running lsof")
	}

	for _, ls := range strings.Split(string(lsof), "\n") {
		if strings.Contains(ls, "TCP") {
			conns++
		}
	}
	return
}

// ============================================================================
// HELPERS
// ============================================================================

const code200Response = `
	<xml>
		<some />
	</xml>
`

type handlers struct {
	called  bool
	timeout time.Duration
}

func (h *handlers) code200(w http.ResponseWriter, r *http.Request) {
	h.called = true
	w.Header().Add("Content-Type", "application/xml")
	io.WriteString(w, code200Response)
}

func (h *handlers) delay(w http.ResponseWriter, r *http.Request) {
	h.code200(w, r)
	time.Sleep(h.timeout)
}

func newServer(f http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(f))
}

func new200Server() *httptest.Server {
	return newServer((&handlers{}).code200)
}

func newServerWithTimeout(f http.HandlerFunc, d time.Duration) *httptest.Server {
	ts := httptest.NewUnstartedServer(f)
	ts.Config.ReadTimeout = d
	ts.Config.WriteTimeout = d
	ts.Start()
	return ts
}

func bodyClose(res *http.Response) string {
	// drain the body and then close it or it'll leak
	defer res.Body.Close()
	return body(res)
}

func body(res *http.Response) string {
	// drain the body
	body, _ := ioutil.ReadAll(res.Body)
	return string(body)
}

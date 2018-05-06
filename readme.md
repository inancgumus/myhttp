# MyHttp
> Easy to use API to make timeout supported http GET requests in Go.

![Go Report Card](https://goreportcard.com/badge/github.com/inancgumus/myhttp) [![Coverage Status](https://coveralls.io/repos/github/inancgumus/myhttp/badge.svg?branch=master)](https://coveralls.io/github/inancgumus/myhttp?branch=master)

MyHttp is for coders who don't want to write timeout support logic and don't want to deal with heavy APIs, just to make http GET request in Go.

It's battle-tested in production and has tests which even verifies connection leaking issues.

I'd been dealing with this issue myself, I read many documents, and saw that there are no simple and lightweight APIs to do http GET request with timeout logic. So, I created MyHttp in one of my projects and wanted to move it here.

## Installation

```sh
go get github.com/inancgumus/myhttp
```

## Documentation

See [MyHttp GoDoc](https://godoc.org/github.com/inancgumus/myhttp) for the documentation.

## Usage example

### Get

`Get` simply gets the url and returns an `http.Response`.

```go
import (
	"time"
	"github.com/inancgumus/myhttp"
)

mh := myhttp.New(time.Second * 10) // timeout after 10 seconds

res, err := mh.Get("http://www.domain.com/foo")
if err != nil {
	panic(err)
}

// use res here...
```

### WrapGet

`WrapGet` accepts a function in its second argument and runs it after getting the `http.Response` from the url and automatically closes the http.Response.Body, in case you may forget to close it.

```go
import (
	"time"
	"github.com/inancgumus/myhttp"
)

mh := myhttp.New(time.Second * 10) // timeout after 10 seconds

err := mh.WrapGet("http://www.domain.com/foo", func(r *http.Response) error {
	// use res here...
	// MyHttp will automatically close http.Response.Body for you after this func ends.
})

if err != nil {
	panic(err)
}
```

## Author

Inanc Gumus – [@inancgumus](https://twitter.com/inancgumus)

Distributed under the MIT license. See ``LICENSE`` for more information.

## Contributing

1. Fork it (<https://github.com/inancgumus/myhttp/fork>)
2. Create your feature branch (`git checkout -b feature/foo`)
3. Commit your changes (`git commit -am 'add: foo'`)
4. Push to the branch (`git push origin feature/foo`)
5. Create a new Pull Request

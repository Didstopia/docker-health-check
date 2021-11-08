package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()
	app.Version = "0.0.3"
	app.Name = "Health Checker"
	app.Usage = "Hits an endpoint for you.  healthcheck -url=http://localhost/ping"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "url, U",
			Usage:       "the full url to hit (required if hostname is not set)",
			Destination: &url,
		},
		cli.StringFlag{
			Name:        "hostname, N",
			Usage:       "the hostname for the request (required if url is not set)",
			Destination: &hostname,
		},
		cli.IntFlag{
			Name:        "port, P",
			Usage:       "the port for the request (optional)",
			Value:       80,
			Destination: &port,
		},
		cli.StringFlag{
			Name:        "schema, S",
			Usage:       "the schema for the request (optional)",
			Value:       "http",
			Destination: &schema,
		},
		cli.StringFlag{
			Name:        "endpoint, E",
			Usage:       "the endpoint for the request (optional)",
			Value:       "",
			Destination: &endpoint,
		},
		cli.StringSliceFlag{
			Name:  "headers, H",
			Usage: "specify a header and value for the request (optional, -H=key:value)",
		},
		cli.StringFlag{
			Name:        "verb, V",
			Usage:       "the HTTP verb to use (optional)",
			Value:       "GET",
			Destination: &httpVerb,
		},
		cli.IntFlag{
			Name:        "code, C",
			Usage:       "expected response code (optional)",
			Value:       http.StatusOK,
			Destination: &statusCode,
		},
		// http body not supported yet
		// response body checking not supported yet
	}
	app.Action = actionFunc

	app.Run(os.Args)
}

func actionFunc(c *cli.Context) error {
	// Validate that either url or hostname is set
	if len(url) < 0 && len(hostname) < 0 {
		return cli.NewExitError("url or hostname length must be > 0 ", 1)
	}

	// Validate that either url or hostname is set, never both
	if len(url) > 0 && len(hostname) > 0 {
		return cli.NewExitError("specify url or hostname, not both ", 1)
	}

	// Validate valid port when hostname is set
	if len(hostname) > 0 && port <= 0 {
		return cli.NewExitError("hostname specified but port is invalid ", 1)
	}

	// Build a url if hostname is specified
	if len(hostname) > 0 {
		url = fmt.Sprintf("%s://%s:%d%s", schema, hostname, port, endpoint)
	}

	req, err := http.NewRequest(httpVerb, url, nil)

	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	for _, str := range c.StringSlice("headers") {
		kv := strings.Split(str, ":")
		if len(kv) == 2 {
			req.Header.Add(kv[0], kv[1])
		} else {
			return cli.NewExitError("header field must be in the format \"key:value\"", 1)
		}
	}

	req.Close = true

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	if resp != nil {
		defer func(r *http.Response) {
			if r.Body != nil {
				r.Body.Close()
			}
		}(resp)
		if resp.StatusCode != statusCode {
			return cli.NewExitError(fmt.Sprintf("resp code %d didn't match %d", resp.StatusCode, statusCode), 1)
		}
	}
	return nil
}

// globals
var (
	url        string
	hostname   string
	port       int
	schema     string
	endpoint   string
	httpVerb   string
	statusCode int
)

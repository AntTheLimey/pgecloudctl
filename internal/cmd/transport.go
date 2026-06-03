package cmd

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

type verboseTransport struct {
	base http.RoundTripper
}

func (t *verboseTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Fprintf(os.Stderr, "> %s %s\n", req.Method, req.URL)

	if auth := req.Header.Get("Authorization"); auth != "" {
		if len(auth) > 17 {
			auth = auth[:17] + "████"
		}
		fmt.Fprintf(os.Stderr, "> Authorization: %s\n", auth)
	}

	start := time.Now()
	resp, err := t.base.RoundTrip(req)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Fprintf(os.Stderr, "< error: %v (%s)\n", err, elapsed.Round(time.Millisecond))
		return nil, err
	}

	fmt.Fprintf(os.Stderr, "< %d %s (%s)\n", resp.StatusCode, http.StatusText(resp.StatusCode), elapsed.Round(time.Millisecond))
	return resp, nil
}

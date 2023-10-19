package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

var maxRedirects = 10
var timeout = 15
var listenAddr = ":8000"

var headerBlacklist = []string{}
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Public",
	"Proxy-Authenticate",
	"Transfer",
	"Upgrade",
}

var proxy_base_url = "PROXY_BASE_URL"

var helpText = fmt.Sprintf(strings.Replace(`NAME
    corsproxy - adds cors headers to requests

SYNOPSIS
    /        - Shows this message
    /{url}   - Requests {url}

DESCRIPTION
    corsproxy allows requests to be made from any origin by adding cors
    headers. It supports all HTTP methods and headers.

    The following additional headers are added to the proxied request:

        Access-Control-Allow-Origin       - Allows access from all origins
        Access-Control-Expose-Headers     - Allows the browser to access
                                            all headers.
        Access-Control-Allow-Credentials  - Allows the browser to access Credentials
        Access-Control-Allow-Headers      - Allows the browser to access

        X-Request-URL                     - The requested URL
        X-Final-URL                       - The final URL after redirects

    The timeout for requests is %d seconds, and corsproxy will follow up
    to %d redirects.

    You can also use the %s env variable to append a base url to the request, in other words, use it as a transparant proxy.

ABOUT
    Source Code at https://github.com/marcosdiez/corsproxy
    Forked from https://github.com/pgaskin/corsproxy
`, "\t", "    ", -1), timeout, maxRedirects, proxy_base_url)

var client = &http.Client{
	Timeout: time.Second * time.Duration(timeout),
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= maxRedirects {
			return errors.New("stopped after 10 redirects")
		}
		return nil
	},
}

func main() {
	pflag.StringSliceVarP(&headerBlacklist, "header-blacklist", "b", headerBlacklist, "Headers to remove from the request and response")
	pflag.StringVarP(&listenAddr, "addr", "a", listenAddr, "Address to listen on")
	pflag.IntVarP(&timeout, "timeout", "t", timeout, "Request timeout")
	pflag.IntVarP(&maxRedirects, "max-redirects", "r", maxRedirects, "Maximum number of redirects to follow")
	help := pflag.BoolP("help", "h", false, "Show this message")
	pflag.Parse()

	if *help || pflag.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "Usage: corsproxy [OPTIONS]\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		pflag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nRun the server and go to it in a web browser for API documentation.\n")
		os.Exit(1)
	}

	fmt.Printf("Listening on %s\n", listenAddr)

	if os.Getenv(proxy_base_url) != "" {
		fmt.Printf("Using [%s] as the base proxy URL\n", os.Getenv(proxy_base_url))
	}

	err := http.ListenAndServe(listenAddr, http.HandlerFunc(handleCORS))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handleCORS(w http.ResponseWriter, req *http.Request) {

	full_url := fmt.Sprintf("%s%s", os.Getenv(proxy_base_url), req.URL.Path)
	p := strings.TrimLeft(full_url, "/")
	// fmt.Printf("Request: %s %s %+v %s\n\n", req.Method, full_url, req.Header, req.Body)

	if p == "" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, helpText)
		return
	}

	if q := req.URL.RawQuery; q != "" {
		p += "?" + q
	}

	// if req.Header.Get("Origin") == "" && req.Header.Get("X-Requested-With") == "" {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	fmt.Fprintf(w, "Error: Origin or X-Requested-With must be specified.")
	// 	return
	// }

	u, err := url.Parse(p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %v", err)
		return
	}

	if u.Scheme == "" {
		u.Scheme = "http"
	}

	nreq, err := http.NewRequest(req.Method, u.String(), req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %v", err)
		return
	}

	for key := range req.Header {
		for _, ignore := range append(hopHeaders, headerBlacklist...) {
			if key == ignore {
				continue
			}
		}
		nreq.Header.Set(key, req.Header.Get(key))
	}

	resp, err := client.Do(nreq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %v", err)
		return
	}
	defer resp.Body.Close()
	expose := []string{"X-Request-URL"}
	for key := range resp.Header {
		for _, ignore := range append(hopHeaders, headerBlacklist...) {
			if key == ignore {
				continue
			}
		}
		w.Header().Set(key, resp.Header.Get(key))
		expose = append(expose, key)
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Expose-Headers", strings.Join(expose, ","))
	w.Header().Set("X-Request-URL", u.String())
	w.Header().Set("X-Final-URL", resp.Request.URL.String())
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	fmt.Printf("%s %s -> %d\n", req.Method, full_url, resp.StatusCode)
	// fmt.Printf("Response: %d %+v %s\n\n\n", resp.StatusCode, w.Header(), resp.Body)
}

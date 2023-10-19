# corsproxy
Proxies requests and adds CORS headers.

````
Usage: corsproxy [OPTIONS]

Options:
  -a, --addr string                Address to listen on (default ":8000")
  -b, --header-blacklist strings   Headers to remove from the request and response
  -h, --help                       Show this message
  -r, --max-redirects int          Maximum number of redirects to follow (default 10)
  -t, --timeout int                Request timeout (default 15)

Run the server and go to it in a web browser for API documentation.

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

    You can also use the PROXY_BASE_URL env variable to append a base url to the request, in other words, use it as a transparant proxy.

ABOUT
    Source Code at https://github.com/marcosdiez/corsproxy
    Forked from https://github.com/pgaskin/corsproxy

````
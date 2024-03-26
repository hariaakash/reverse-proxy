package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

const HTTP_ADDR string = ":8080"

func createReverseProxy(c *gin.Context) {
	// Target
	target := "https://jsonplaceholder.typicode.com/"
	// Parse the target URL
	targetURL, err := url.Parse(target)
	if err != nil {
		panic(err)
	}

	// Create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = targetURL.Host
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.URL.Path = c.Param("proxyPath")
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func main() {
	fmt.Println("Hello, World!")

	r := gin.Default()

	r.Any("/*proxyPath", createReverseProxy)

	r.Run()
}
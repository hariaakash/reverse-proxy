package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const HTTP_ADDR string = ":8080"

func handleWebProxy(targetURL *url.URL, c *gin.Context) {
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

func createProxyHandler(c *gin.Context) {
	isWebSocket := c.Request.Header.Get("Upgrade") == "websocket"
	target := "https://jsonplaceholder.typicode.com/"
	if isWebSocket {
		target = "wss://echo.websocket.org/"
	}
	// Parse the target URL
	targetURL, err := url.Parse(target)
	if err != nil {
		handleError(c, err)
		return
	}

	// Check if it's a WebSocket request
	if isWebSocket {
		handleWebSocketProxy(targetURL, c)
		return
	}

	handleWebProxy(targetURL, c)
}

func handleWebSocketProxy(targetURL *url.URL, c *gin.Context) {
	// Upgrade the HTTP connection to WebSocket
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	connBackend, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		handleError(c, fmt.Errorf("error upgrading to WebSocket: %v", err))
		return
	}
	defer connBackend.Close()

	// Connect to WebSocket backend
	connBackendBackend, _, err := websocket.DefaultDialer.Dial(targetURL.String(), nil)
	if err != nil {
		handleError(c, fmt.Errorf("error connecting to WebSocket backend: %v", err))
		return
	}
	defer connBackendBackend.Close()

	// Forward WebSocket messages bidirectionally
	go func() {
		defer connBackendBackend.Close()
		for {
			messageType, message, err := connBackend.ReadMessage()
			if err != nil {
				handleError(c, fmt.Errorf("error reading message from client: %v", err))
				return
			}
			err = connBackendBackend.WriteMessage(messageType, message)
			if err != nil {
				handleError(c, fmt.Errorf("error writing message to backend: %v", err))
				return
			}
		}
	}()

	for {
		messageType, message, err := connBackendBackend.ReadMessage()
		if err != nil {
			handleError(c, fmt.Errorf("error reading message from backend: %v", err))
			return
		}
		err = connBackend.WriteMessage(messageType, message)
		if err != nil {
			handleError(c, fmt.Errorf("error writing message to client: %v", err))
			return
		}
	}
}

func handleError(c *gin.Context, err error) {
	fmt.Println("Error:", err)
	c.AbortWithStatus(http.StatusInternalServerError)
}

func main() {
	fmt.Println("Hello, World!")

	r := gin.Default()

	r.Any("/*proxyPath", createProxyHandler)

	r.Run(HTTP_ADDR)
}

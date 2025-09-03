package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

const limitDeltaTime = 800 * time.Millisecond

type RateLimiter struct {
	clients map[string]time.Time
}

func NewRateLimiter() RateLimiter {
	return RateLimiter{
		clients: make(map[string]time.Time),
	}
}

func (rl *RateLimiter) IsWithinLimit(ip string) bool {
	now := time.Now()
	if last, exists := rl.clients[ip]; exists {
		if now.Sub(last) < limitDeltaTime {
			return false
		}
	}
	rl.clients[ip] = now
	return true
}

var serverPort int
var serverAddress net.IP
var rateLimiter RateLimiter

func loadConsts() {
	serverPort = 8080
	serverAddress = net.IPv4(127, 0, 0, 1)
	rateLimiter = NewRateLimiter()
}

func main() {
	loadConsts()

	apiUrl := os.Args[1]
	target, _ := url.Parse(apiUrl)
	proxy := httputil.NewSingleHostReverseProxy(target)

	serverUrl := fmt.Sprintf("%s:%d", serverAddress, serverPort)
	log.Printf("Garbanzo listening to `%s`", apiUrl)
	log.Printf("Proxy Available at `http://%s:%d`", serverAddress, serverPort)

	log.Fatal(
		http.ListenAndServe(serverUrl, http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				r.Header.Set("X-Proxy", "Garbanzo")

				incommingIp, _, _ := net.SplitHostPort(r.RemoteAddr)
				log.Printf("Request from `%s` -> `%s`", incommingIp, r.URL.Path)

				if !rateLimiter.IsWithinLimit(incommingIp) {
					http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
					return
				}

				proxy.ServeHTTP(w, r)
			},
		)),
	)

}

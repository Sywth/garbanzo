package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var serverPort int
var serverAddress net.IP
var rateLimiter RateLimiter

type SrcExtractor func(*http.Request) (string, error)

var srcExtractorStrategy SrcExtractor = srcAddrExtractorViaRemoteAddr

// Default fallback strategy, just use RemoteAddr
func srcAddrExtractorViaRemoteAddr(r *http.Request) (string, error) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		panic(err)
	}
	return ip, nil
}

// Set of remote addresses we trust to provide valid X-Forwarded-For headers
// (e.g. our load balancers, cloudflare, etc)
var trustedRemoteAddrsForXff map[string]struct{}

func srcAddrExtractorViaXff(r *http.Request) (string, error) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		panic(err)
	}

	if _, ok := trustedRemoteAddrsForXff[ip]; !ok {
		return srcAddrExtractorViaRemoteAddr(r)
	}
	xff := r.Header.Get("X-Forwarded-For")
	if xff == "" {
		return srcAddrExtractorViaRemoteAddr(r)
	}
	return xff, nil
}

func setConsts() {
	serverPort = 8080
	serverAddress = net.IPv4(127, 0, 0, 1)
	rateLimiter = NewRateLimiter(800 * time.Second)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	trustedRemoteAddrsForXff = map[string]struct{}{
		"127.0.0.1": {},
	}
	srcExtractorStrategy = srcAddrExtractorViaXff
}

func rateLimitHandler(w http.ResponseWriter, r *http.Request, proxy *httputil.ReverseProxy) {
	incommingIp, err := srcExtractorStrategy(r)
	log.Debug().Msgf("Request from %s -> %s", incommingIp, r.URL.Path)

	if err != nil {
		log.Info().Msgf("Could not parse IP address `%s` due to `%s`", r.RemoteAddr, err)
		http.Error(w, "Could not parse IP address", http.StatusInternalServerError)
		return
	}

	if !rateLimiter.IsWithinLimit(incommingIp) {
		log.Info().Msgf("Rate limited %s", incommingIp)
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	r.Header.Set("X-Proxy", "Garbanzo")
	proxy.ServeHTTP(w, r)
}

func main() {
	setConsts()

	apiUrl := os.Args[1]
	target, _ := url.Parse(apiUrl)
	proxy := httputil.NewSingleHostReverseProxy(target)

	serverAddr := fmt.Sprintf("%s:%d", serverAddress, serverPort)
	serverUrl := fmt.Sprintf("http://%s", serverAddr)
	log.Debug().Msgf("Garbanzo listening to       %s", apiUrl)
	log.Debug().Msgf("Garbanzo Proxy Available at %s", serverUrl)

	err := http.ListenAndServe(serverAddr, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			rateLimitHandler(w, r, proxy)
		},
	))

	if err != nil {
		log.Fatal().Msgf("Could not start Garbanzo Proxy due to `%s`", err)
	}
}

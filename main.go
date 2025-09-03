package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	target, _ := url.Parse("http://127.0.0.1:8000") // FastAPI app
	proxy := httputil.NewSingleHostReverseProxy(target)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Optional: manipulate headers before forwarding
		r.Header.Set("X-Proxy", "GoReverseProxy")
		proxy.ServeHTTP(w, r)
	})

	log.Println("Reverse proxy listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

package main

import (
	"bufio"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
)

var (
	routes   = map[string]string{}
	routesMu sync.RWMutex
)

func loadRoutes() {
	file, err := os.Open("/srv/router/routes.txt")
	if err != nil {
		log.Println("routes file not found yet")
		return
	}
	defer file.Close()

	newRoutes := map[string]string{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "=")
		if len(parts) == 2 {
			newRoutes[parts[0]] = parts[1]
		}
	}

	routesMu.Lock()
	routes = newRoutes
	routesMu.Unlock()

	log.Println("Routes loaded:", routes)
}

func main() {
	loadRoutes()

	http.HandleFunc("/__reload", func(w http.ResponseWriter, r *http.Request) {
		loadRoutes()
		w.Write([]byte("reloaded"))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := strings.Split(r.Host, ":")[0]

		routesMu.RLock()
		port, ok := routes[host]
		routesMu.RUnlock()

		if !ok {
			http.Error(w, "unknown app", 404)
			return
		}

		target, _ := url.Parse("http://localhost:" + port)
		proxy := httputil.NewSingleHostReverseProxy(target)

		proxy.Transport = &http.Transport{
			DisableKeepAlives: true,
		}

		proxy.ServeHTTP(w, r)
	})

	log.Println("🌐 Router running on :9001")
	log.Fatal(http.ListenAndServe(":9001", nil))
}

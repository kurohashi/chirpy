package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func apiMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		println("API:", r.Method, r.URL.Path)

		// future:
		// auth
		// rate limit
		// headers

		next.ServeHTTP(w, r)
	})
}

func main() {
	serveMux := http.NewServeMux()
	routes(serveMux)
	server := http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}
	server.ListenAndServe()
}

func routes(mux *http.ServeMux) {
	hitConfig := apiConfig{}
	mux.Handle("/app/", hitConfig.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	// mount api router
	mux.Handle("/admin/", http.StripPrefix("/admin", newAdminRouter(&hitConfig)))
	mux.Handle("/api/", http.StripPrefix("/api", newAPIRouter(&hitConfig)))
}

func newAPIRouter(hitConfig *apiConfig) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", healthHandler)

	return apiMiddleware(mux) // attach middleware once
}

func newAdminRouter(hitConfig *apiConfig) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) { getMetrics(w, r, hitConfig) })
	mux.HandleFunc(("POST /reset"), func(w http.ResponseWriter, r *http.Request) { resetMetrics(w, r, hitConfig) })

	return apiMiddleware(mux) // attach middleware once
}

func resetMetrics(res http.ResponseWriter, req *http.Request, config *apiConfig) {
	config.fileserverHits.Store(0)
	res.WriteHeader(200)
	res.Write([]byte{})
}

func getMetrics(res http.ResponseWriter, req *http.Request, config *apiConfig) {
	res.Header().Set("Content-Type", "text/html")
	res.WriteHeader(200)
	response := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", config.fileserverHits.Load())
	res.Write([]byte(response))
}

func healthHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(200)
	res.Write([]byte("OK"))
}

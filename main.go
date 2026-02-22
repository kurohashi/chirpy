package main

import (
	"net/http"
	"strconv"
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
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) { getMetrics(w, r, &hitConfig) })
	mux.HandleFunc(("POST /reset"), func(w http.ResponseWriter, r *http.Request) { resetMetrics(w, r, &hitConfig) })
	mux.HandleFunc("GET /healthz", healthHandler)
}

func resetMetrics(res http.ResponseWriter, req *http.Request, config *apiConfig) {
	config.fileserverHits.Store(0)
	res.WriteHeader(200)
	res.Write([]byte{})
}

func getMetrics(res http.ResponseWriter, req *http.Request, config *apiConfig) {
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(200)
	response := "Hits: " + strconv.Itoa(int(config.fileserverHits.Load()))
	res.Write([]byte(response))
}

func healthHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(200)
	res.Write([]byte("OK"))
}

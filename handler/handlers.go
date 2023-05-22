package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/shokHorizon/ping_api/entity"
)

func GetLatencyHandler(m map[string]int64) http.HandlerFunc {
	var counter int64 = 0
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if metrics := q.Get("m"); metrics != "" {
			io.WriteString(w, fmt.Sprintf("Counter: %d", counter))
			return
		}
		counter += 1
		if url := q.Get("q"); url != "" {
			if val, ok := m[url]; ok {
				io.WriteString(w, fmt.Sprintf("Latency: %dms", val))
				return
			}
		}
		io.WriteString(w, "Invalid link")
	}
}

func GetMaxLatencyHandler(kv *entity.KeyVal) http.HandlerFunc {
	var counter int64 = 0
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if metrics := q.Get("m"); metrics != "" {
			io.WriteString(w, fmt.Sprintf("Counter: %d", counter))
			return
		}
		counter += 1
		io.WriteString(w, fmt.Sprintf("%s has max latency: %dms", kv.Key, kv.Val))
	}
}

func GetMinLatencyHandler(kv *entity.KeyVal) http.HandlerFunc {
	var counter int64 = 0
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if metrics := q.Get("m"); metrics != "" {
			io.WriteString(w, fmt.Sprintf("Counter: %d", counter))
			return
		}
		counter += 1
		io.WriteString(w, fmt.Sprintf("%s has min latency: %dms", kv.Key, kv.Val))
	}
}

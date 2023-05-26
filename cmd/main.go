package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/shokHorizon/ping_api/entity"
	"github.com/shokHorizon/ping_api/handler"
)

const (
	max_pingers = 10
)

func scedule(ctx context.Context, c chan<- string, t time.Duration, urls []string) {
	done := false

	for !done {
		select {
		case <-time.After(t):
			for _, url := range urls {
				c <- url
			}
		case <-ctx.Done():
			done = true
			close(c)
		}
	}
}

func ping(ctx context.Context, taskc <-chan string, respc chan<- entity.KeyVal) {
	done := false

	var url string
	var t time.Time

	for !done {
		select {
		case url = <-taskc:
			log.Printf("Got task: %s\n", url)
			t = time.Now()
			r, err := http.Get("https://" + url)
			if err != nil {
				log.Println(err)
				continue
			}
			r.Body.Close()
			respc <- entity.KeyVal{Key: url, Val: time.Since(t).Milliseconds()}
			log.Printf("Response %s took %d mils\n", url, time.Since(t).Milliseconds())
		case <-ctx.Done():
			done = true
		}
	}

}

func record(ctx context.Context, respc <-chan entity.KeyVal, m map[string]int64, minLat, maxLat *entity.KeyVal) {
	done := false

	var kv entity.KeyVal

	for !done {
		select {
		case kv = <-respc:
			m[kv.Key] = kv.Val
			if kv.Val < minLat.Val {
				*minLat = kv
			} else if kv.Val > maxLat.Val {
				*maxLat = kv
			} else {
				var newMin entity.KeyVal
				var newMax entity.KeyVal
				for k, v := range m {
					if newMin.Val == 0 {
						newMin = entity.KeyVal{Key: k, Val: v}
						newMax = newMin
						continue
					}
					if v < newMin.Val {
						newMin.Key = k
						newMin.Val = v
					} else if v > newMax.Val {
						newMax.Key = k
						newMax.Val = v
					}
				}
				*minLat = newMin
				*maxLat = newMax
			}
		case <-ctx.Done():
			done = true
		}
	}
}

func parseUrls() ([]string, error) {
	if len(os.Args) < 1 {
		log.Fatal("Specify the urls path")
	}
	path := os.Args[1]
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(bytes), "\n"), nil
}

func main() {
	urls, err := parseUrls()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Got %d links\n", len(urls))

	wg := sync.WaitGroup{}
	m := make(map[string]int64)
	ctx, cancel := context.WithCancel(context.Background())
	tasks := make(chan string, max_pingers)
	resps := make(chan entity.KeyVal, max_pingers)
	minLat := entity.KeyVal{}
	maxLat := entity.KeyVal{}

	r := http.NewServeMux()
	r.HandleFunc("/latency", handler.GetLatencyHandler(m))
	r.HandleFunc("/min", handler.GetMinLatencyHandler(&minLat))
	r.HandleFunc("/max", handler.GetMaxLatencyHandler(&maxLat))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("Received interrupt signal, cancelling context...")
		cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server Shutdown Failed:%+v", err)
		}

	}()

	for i := 0; i < max_pingers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ping(ctx, tasks, resps)
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		scedule(ctx, tasks, time.Minute, urls)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		record(ctx, resps, m, &minLat, &maxLat)
	}()

	err = srv.ListenAndServe()
	if err != nil {
		log.Println(err)
	}

}

package loadbalancergo

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

type ServerPool struct {
	backends []*Backend
	current  uint64
}

func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

func (b *Backend) IsAlive() (alive bool) {
	b.mux.RLock()
	alive = b.Alive
	b.mux.Unlock()
	return
}

func (s *ServerPool) GetNextPeer() *Backend {
	next := s.NextIndex()

	l := len(s.backends) + next

	for i := next; i < l; i++ {
		idx := i % len(s.backends)

		if s.backends[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx))
			}

			return s.backends[idx]
		}
	}
	return nil
}

func lb(w http.ResponseWriter, r *http.Request) {

	peer := serverPool.GetNextPeer()

	if peer != nil {
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}

	http.Error(w, "Service not available", http.StatusServiceUnavailable)

}

var serverPool ServerPool

func main() {
	u, _ := url.Parse("http://localhost:8080")

	rp := httputil.NewSingleHostReverseProxy(u)

	http.HandlerFunc(rp.ServeHTTP)

	server := http.Server{
		Addr:    fmt.Sprintf(":d", port),
		Handler: http.HandlerFunc(lb),
	}

	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {

		log.Printf("[%s] %s\n", serverUrl.Host, e.Error())

		retries := GetRetryFromContext(request)

		if retries < 3 {
			select {

			case <-time.After(10 * time.Millisecond):
				ctx := context.WithValue(request.Context(), Retry, retries+1)
				proxy.ServeHTTP(writer, request.WithContext(ctx))
			}

			return
		}
	}

	ServerPool.MarkBackendStatus(serverUrl, false)

	attempts := GetAttemptsFromContext(request)

}

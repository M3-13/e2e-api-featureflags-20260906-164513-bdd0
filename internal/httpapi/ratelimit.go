package httpapi

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// windowEntry hält den Fenster-Start und die Anzahl der Requests eines
// einzelnen Clients innerhalb des Sliding-Window.
type windowEntry struct {
	start time.Time
	count int
}

// limiter hält den Zustand des Sliding-Window-Rate-Limiters pro Client-IP.
type limiter struct {
	mu      sync.Mutex
	window  time.Duration
	rps     int
	clients map[string]*windowEntry
}

func newLimiter(rps int, window time.Duration) *limiter {
	return &limiter{
		window:  window,
		rps:     rps,
		clients: make(map[string]*windowEntry),
	}
}

// allow meldet, ob der Client innerhalb des aktuellen Fensters noch einen
// Request ausführen darf, und zählt ihn andernfalls.
func (l *limiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	e, ok := l.clients[ip]
	if !ok || now.Sub(e.start) >= l.window {
		l.clients[ip] = &windowEntry{start: now, count: 1}
		return true
	}
	if e.count >= l.rps {
		return false
	}
	e.count++
	return true
}

// RateLimit begrenzt die Anzahl der Requests pro Sekunde pro Client-IP über
// ein Sliding-Window von einer Sekunde. Überschreitungen werden mit 429
// beantwortet. Bei rps <= 0 wird nicht limitiert.
func RateLimit(next http.Handler, rps int) http.Handler {
	if rps <= 0 {
		return next
	}
	l := newLimiter(rps, time.Second)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !l.allow(ip) {
			writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientIP ermittelt die Client-IP aus der RemoteAddr. X-Forwarded-For wird
// bewusst ignoriert, da die Terminierung am Reverse-Proxy erfolgt und dort
// kein vertrauenswürdiger Header eingespeist wird.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

package server

import (
	"expvar"
	"net/http"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mijia/sweb/log"
	"github.com/paulbellamy/ratecounter"
	"golang.org/x/net/context"
)

type LatencyCounter struct {
	sync.RWMutex
	latency []int64 // in nanoseconds
	size    int
}

type LatencyStat struct {
	Latencies []string
	Max       string
	Min       string
	Average   string
	LatP95    string
	LatP75    string
	LatP50    string
}

func (lc *LatencyCounter) Add(latency time.Duration) {
	lc.Lock()
	defer lc.Unlock()
	newSlice := make([]int64, 1+len(lc.latency))
	newSlice[0] = int64(latency / time.Nanosecond)
	copy(newSlice[1:], lc.latency)
	lc.latency = newSlice
	if len(lc.latency) > lc.size {
		lc.latency = lc.latency[:lc.size]
	}
}

func (lc *LatencyCounter) Stat() LatencyStat {
	lc.RLock()
	latency := make([]int64, len(lc.latency))
	copy(latency, lc.latency)
	lc.RUnlock()

	stat := LatencyStat{}
	stat.Latencies = make([]string, len(latency))
	var sum int64
	for i, lat := range latency {
		stat.Latencies[i] = time.Duration(lat).String()
		sum += lat
	}
	if len(latency) > 0 {
		stat.Average = time.Duration(sum / int64(len(latency))).String()
		sort.Sort(int64Slice(latency))
		stat.Max = time.Duration(latency[len(latency)-1]).String()
		stat.Min = time.Duration(latency[0]).String()
		stat.LatP95 = time.Duration(latency[len(latency)*95/100]).String()
		stat.LatP75 = time.Duration(latency[len(latency)*75/100]).String()
		stat.LatP50 = time.Duration(latency[len(latency)/2]).String()
	}
	return stat
}

func NewLatencyCounter(size int) *LatencyCounter {
	return &LatencyCounter{
		latency: make([]int64, 0, size),
		size:    size,
	}
}

// RuntimeWare is the statictics middleware which would collect some basic qps, 4xx, 5xx data information
type RuntimeWare struct {
	serverStarted time.Time
	trackPageview bool
	ignoredUrls   []string

	cQps *ratecounter.RateCounter
	c4xx *ratecounter.RateCounter
	c5xx *ratecounter.RateCounter
	lc   *LatencyCounter

	hitsTotal    *expvar.Int
	hitsQps      *expvar.Int
	hits4xx      *expvar.Int
	hits5xx      *expvar.Int
	hitsServed   *expvar.String
	hitsLatMax   *expvar.String
	hitsLatMin   *expvar.String
	hitsLat95    *expvar.String
	hitsLat50    *expvar.String
	numGoroutine *expvar.Int
	pageviews    *expvar.Map
}

func (m *RuntimeWare) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request, next Handler) context.Context {
	start := time.Now()
	newCtx := next(ctx, w, r)

	urlPath := r.URL.Path
	statusCode := w.(ResponseWriter).Status()
	if statusCode >= 500 {
		m.c5xx.Incr(1)
		m.hits5xx.Set(m.c5xx.Rate())
	} else if statusCode >= 400 {
		m.c4xx.Incr(1)
		m.hits4xx.Set(m.c4xx.Rate())
	}

	ignoreQps := false
	for _, prefix := range m.ignoredUrls {
		if strings.HasPrefix(urlPath, prefix) {
			ignoreQps = true
			break
		}
	}
	if !ignoreQps {
		m.cQps.Incr(1)
		rate := m.cQps.Rate()
		m.hitsQps.Set(rate)
		m.hitsTotal.Add(1)
		m.lc.Add(time.Since(start))
		lcStat := m.lc.Stat()
		m.hitsServed.Set(strings.Join(lcStat.Latencies, ","))
		m.hitsLatMax.Set(lcStat.Max)
		m.hitsLatMin.Set(lcStat.Min)
		m.hitsLat95.Set(lcStat.LatP95)
		m.hitsLat50.Set(lcStat.LatP50)

		if m.trackPageview {
			m.pageviews.Add(urlPath, 1)
		}
	}
	m.numGoroutine.Set(int64(runtime.NumGoroutine()))
	return newCtx
}

func (m *RuntimeWare) logSnapshot(interval time.Duration) {
	c := time.Tick(interval)
	for _ = range c {
		log.Infof("[RuntimeWare] Snapshot server_lived=%s, hits_total=%s, hits_1m=%s, 4xx_5m=%s, hits_5xx=%s, num_goroutine=%s, latency=%s",
			time.Since(m.serverStarted),
			m.hitsTotal, m.hitsQps, m.hits4xx, m.hits5xx, m.numGoroutine, m.hitsServed)
	}
}

func NewRuntimeWare(prefixes []string, trackPageview bool, logInterval ...time.Duration) Middleware {
	expvar.NewString("at_server_start").Set(time.Now().Format("2006-01-02 15:04:05"))
	expvar.NewInt("cpu_count").Set(int64(runtime.NumCPU()))
	ware := &RuntimeWare{
		serverStarted: time.Now(),
		trackPageview: trackPageview,
		ignoredUrls:   prefixes,
		cQps:          ratecounter.NewRateCounter(time.Minute),
		c4xx:          ratecounter.NewRateCounter(5 * time.Minute),
		c5xx:          ratecounter.NewRateCounter(5 * time.Minute),
		lc:            NewLatencyCounter(50),
		hitsTotal:     expvar.NewInt("hits_total"),
		hitsQps:       expvar.NewInt("hits_per_minute"),
		hits4xx:       expvar.NewInt("hits_4xx_per_5min"),
		hits5xx:       expvar.NewInt("hits_5xx_per_5min"),
		hitsServed:    expvar.NewString("latency_recent"),
		hitsLatMax:    expvar.NewString("latency_max"),
		hitsLatMin:    expvar.NewString("latency_min"),
		hitsLat95:     expvar.NewString("latency_p95"),
		hitsLat50:     expvar.NewString("latency_p50"),
		numGoroutine:  expvar.NewInt("goroutine_count"),
	}
	if trackPageview {
		ware.pageviews = expvar.NewMap("hits_pageviews")
	}
	if len(logInterval) > 0 && logInterval[0] > 0 {
		go ware.logSnapshot(logInterval[0])
	}
	return ware
}

type int64Slice []int64

func (p int64Slice) Len() int           { return len(p) }
func (p int64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p int64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

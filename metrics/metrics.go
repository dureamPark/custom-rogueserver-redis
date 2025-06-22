// metrics/metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
)

var (
    CacheHits = prometheus.NewCounter(
        prometheus.CounterOpts{
            Namespace: "pokerogue",           // 프로젝트명 네임스페이스
            Subsystem: "session_cache",
            Name:      "hits_total",
            Help:      "Number of Redis session cache hits",
        },
    )
    CacheMisses = prometheus.NewCounter(
        prometheus.CounterOpts{
            Namespace: "pokerogue",
            Subsystem: "session_cache",
            Name:      "misses_total",
            Help:      "Number of Redis session cache misses",
        },
    )
)

func init() {
    // 애플리케이션 구동 시 자동으로 메트릭을 등록
    prometheus.MustRegister(CacheHits)
    prometheus.MustRegister(CacheMisses)
}


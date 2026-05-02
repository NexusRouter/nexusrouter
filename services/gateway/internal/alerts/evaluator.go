// Package alerts 根据指标与 admin_alerts 配置评估管理面板告警状态。
package alerts

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/metrics"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"go.uber.org/zap"
)

// Status 对外展示级别。
type Status struct {
	Level   string   `json:"level"`   // ok | warning | critical
	Reasons []string `json:"reasons"` // 编码，如 HIGH_ERROR_RATE
	Enabled bool     `json:"enabled"`
}

var (
	mu          sync.RWMutex
	current     = Status{Level: "ok", Reasons: nil, Enabled: false}
	startOnce   sync.Once
	tickDur     = 15 * time.Second
	minTotal    = uint64(40)
	consecutive uint32
)

// Current 返回最近一次评估结果（线程安全）。
func Current() Status {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

func setStatus(s Status) {
	mu.Lock()
	current = s
	mu.Unlock()
}

func numAsUint64(v any) uint64 {
	switch x := v.(type) {
	case uint64:
		return x
	case uint:
		return uint64(x)
	case int:
		if x > 0 {
			return uint64(x)
		}
	case float64:
		if x > 0 {
			return uint64(x)
		}
	}
	return 0
}

// StartBackground 启动后台评估循环（进程级单例）。
func StartBackground(rt *runtime.Store, col *metrics.Collector, log *zap.Logger) {
	if rt == nil || col == nil {
		return
	}
	startOnce.Do(func() {
		go loop(context.Background(), rt, col, log)
	})
}

func loop(ctx context.Context, rt *runtime.Store, col *metrics.Collector, log *zap.Logger) {
	t := time.NewTicker(tickDur)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			evalOnce(rt, col, log)
		}
	}
}

func evalOnce(rt *runtime.Store, col *metrics.Collector, log *zap.Logger) {
	snap := rt.Snapshot()
	al := snap.AdminAlerts
	enabled := al.Enabled && al.ErrorRateThreshold > 0
	if !enabled {
		atomic.StoreUint32(&consecutive, 0)
		setStatus(Status{Level: "ok", Reasons: nil, Enabled: false})
		return
	}
	consNeed := al.ConsecutivePeriods
	if consNeed < 1 {
		consNeed = 2
	}
	sum := col.SummaryJSON()
	total := numAsUint64(sum["requests_total"])
	sr, _ := sum["success_rate"].(float64)
	if total < minTotal {
		atomic.StoreUint32(&consecutive, 0)
		setStatus(Status{Level: "ok", Reasons: nil, Enabled: true})
		return
	}
	errRate := 1.0 - sr
	if errRate < 0 {
		errRate = 0
	}
	if errRate >= al.ErrorRateThreshold {
		n := atomic.AddUint32(&consecutive, 1)
		if uint32(consNeed) <= n {
			setStatus(Status{Level: "critical", Reasons: []string{"HIGH_ERROR_RATE"}, Enabled: true})
			if log != nil {
				log.Warn("admin alert critical", zap.Float64("error_rate", errRate), zap.Float64("threshold", al.ErrorRateThreshold))
			}
			return
		}
		setStatus(Status{Level: "warning", Reasons: []string{"HIGH_ERROR_RATE"}, Enabled: true})
		return
	}
	atomic.StoreUint32(&consecutive, 0)
	setStatus(Status{Level: "ok", Reasons: nil, Enabled: true})
}

package auth

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter IPアドレスベースのレート制限
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	// 1秒あたりのリクエスト数
	rps float64
	// バースト許容量
	burst int
	// クリーンアップ間隔
	cleanupInterval time.Duration
}

// NewRateLimiter レート制限インスタンスを作成
// rps: 1秒あたりの許可リクエスト数
// burst: 瞬間的に許可するリクエスト数
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		limiters:        make(map[string]*rate.Limiter),
		rps:             rps,
		burst:           burst,
		cleanupInterval: 5 * time.Minute,
	}

	// 定期的に古いエントリをクリーンアップ
	go rl.cleanupLoop()

	return rl
}

// GetLimiter IPアドレスのリミッターを取得または作成
func (rl *RateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(rl.rps), rl.burst)
		rl.limiters[ip] = limiter
	}

	return limiter
}

// cleanupLoop 定期的に未使用のリミッターを削除
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		// 簡易実装: 全てクリア（本番ではLRUキャッシュ等を検討）
		rl.limiters = make(map[string]*rate.Limiter)
		rl.mu.Unlock()
	}
}

// Allow リクエストを許可するか判定
func (rl *RateLimiter) Allow(ip string) bool {
	limiter := rl.GetLimiter(ip)
	return limiter.Allow()
}

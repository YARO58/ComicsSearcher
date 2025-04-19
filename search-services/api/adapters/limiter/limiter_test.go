package limiter

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConcurrencyLimiter(t *testing.T) {
	tests := []struct {
		name     string
		limit    int
		acquires int
		expected int
	}{
		{
			name:     "single acquire",
			limit:    1,
			acquires: 1,
			expected: 1,
		},
		{
			name:     "multiple acquires within limit",
			limit:    3,
			acquires: 2,
			expected: 2,
		},
		{
			name:     "acquires exceeding limit",
			limit:    2,
			acquires: 3,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewConcurrencyLimiter(tt.limit)
			successfulAcquires := 0

			for i := 0; i < tt.acquires; i++ {
				if limiter.Acquire() {
					successfulAcquires++
				}
			}

			assert.Equal(t, tt.expected, successfulAcquires)

			// Освобождаем все семафоры
			for i := 0; i < successfulAcquires; i++ {
				limiter.Release()
			}
		})
	}
}

func TestConcurrencyLimiter_Concurrent(t *testing.T) {
	limiter := NewConcurrencyLimiter(3)
	var wg sync.WaitGroup
	successfulAcquires := 0
	var mu sync.Mutex

	// Запускаем 10 горутин, пытающихся получить семафор
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if limiter.Acquire() {
				mu.Lock()
				successfulAcquires++
				mu.Unlock()
				time.Sleep(10 * time.Millisecond) // Имитируем работу
				limiter.Release()
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, 3, successfulAcquires)
}

func TestRateLimiter(t *testing.T) {
	tests := []struct {
		name     string
		rps      float64
		attempts int
		expected int
	}{
		{
			name:     "low rate",
			rps:      1.0,
			attempts: 3,
			expected: 1,
		},
		{
			name:     "high rate",
			rps:      10.0,
			attempts: 5,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewRateLimiter(tt.rps)
			successfulAttempts := 0

			for i := 0; i < tt.attempts; i++ {
				if limiter.Allow() {
					successfulAttempts++
				}
				time.Sleep(10 * time.Millisecond) // Даем время для обновления лимитера
			}

			assert.Equal(t, tt.expected, successfulAttempts)
		})
	}
}

func TestRateLimiter_Wait(t *testing.T) {
	limiter := NewRateLimiter(1.0) // 1 запрос в секунду
	start := time.Now()

	// Делаем 3 запроса
	for i := 0; i < 3; i++ {
		limiter.Wait()
	}

	duration := time.Since(start)
	assert.True(t, duration >= 2*time.Second, "Должно пройти минимум 2 секунды между запросами")
}

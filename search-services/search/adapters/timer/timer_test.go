package timer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockUpdater struct {
	updateCount int
	updateErr   error
}

func (m *mockUpdater) UpdateIndex(ctx context.Context) {
	m.updateCount++
}

func TestTimer_Start(t *testing.T) {
	tests := []struct {
		name        string
		seconds     int
		updateErr   error
		ctxDuration time.Duration
		minExpected int
	}{
		{
			name:        "successful updates",
			seconds:     1,
			updateErr:   nil,
			ctxDuration: 3 * time.Second,
			minExpected: 3, // Минимум 3 обновления (начальное + 2 по таймеру)
		},
		{
			name:        "update error",
			seconds:     1,
			updateErr:   errors.New("update error"),
			ctxDuration: 3 * time.Second,
			minExpected: 3, // Минимум 3 обновления (начальное + 2 по таймеру)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updater := &mockUpdater{
				updateErr: tt.updateErr,
			}
			timer := New(tt.seconds, updater)

			ctx, cancel := context.WithTimeout(context.Background(), tt.ctxDuration)
			defer cancel()

			timer.Start(ctx)

			// Даем немного времени на завершение всех горутин
			time.Sleep(100 * time.Millisecond)

			assert.GreaterOrEqual(t, updater.updateCount, tt.minExpected)
		})
	}
}

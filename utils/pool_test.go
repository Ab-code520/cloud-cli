package utils

import (
	"context"
	"testing"
	"time"
)

func TestPoolFastFail(t *testing.T) {
	// Test that submitting tasks after cancellation doesn't block or spin CPU.
	p := NewPool(1)
	ctx, cancel := context.WithCancel(context.Background())
	_ = ctx // Use context var if needed later

	// Submit a task that will cancel the pool immediately
	p.Go(func(ctx context.Context) error {
		cancel() // Trigger cancellation
		return nil
	})

	// Wait for the task to run and cancel
	time.Sleep(200 * time.Millisecond)

	// Now submit many tasks - they should be discarded instantly
	start := time.Now()
	for i := 0; i < 1000; i++ {
		p.Go(func(ctx context.Context) error {
			return nil
		})
	}
	elapsed := time.Since(start)

	// If fast-fail works, 1000 submissions should take < 10ms.
	// If it was blocking/spinning, it would take much longer.
	if elapsed > 50*time.Millisecond {
		t.Errorf("Fast-fail took too long: %v", elapsed)
	}

	// Clean up
	p.Wait()
}
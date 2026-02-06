// Package middleware testa o padrão Circuit Breaker.
//
// Estes testes validam:
// - Transições de estado (CLOSED -> OPEN -> HALF-OPEN -> CLOSED)
// - Contagem de falhas e sucessos
// - Timeout de reset
// - Thread-safety
//
// Tipo: Testes unitários
// Previne: Circuit breaker que não abre, não fecha, ou não respeita timeouts
package middleware

import (
	"sync"
	"testing"
	"time"

	gl "github.com/kubex-ecosystem/logz"
)

func init() {
	opts := gl.NewLogzOptions(false)
	opts.Level = gl.ParseLevel("warn")
	gl.SetGlobalLoggerZ(gl.NewLoggerZ("gnyx-test", opts, false))
}

// --- Testes de CircuitState.String() ---

func TestCircuitState_String(t *testing.T) {
	tests := []struct {
		state CircuitState
		want  string
	}{
		{CircuitClosed, "CLOSED"},
		{CircuitOpen, "OPEN"},
		{CircuitHalfOpen, "HALF-OPEN"},
		{CircuitState(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.state.String()
			if got != tt.want {
				t.Errorf("CircuitState(%d).String() = %q, want %q", tt.state, got, tt.want)
			}
		})
	}
}

// --- Testes de NewCircuitBreaker ---

func TestNewCircuitBreaker(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures:      5,
		ResetTimeout:     time.Second,
		SuccessThreshold: 2,
	}

	cb := NewCircuitBreaker(config)

	if cb == nil {
		t.Fatal("NewCircuitBreaker() returned nil")
	}

	state, failures, successes := cb.GetState()
	if state != CircuitClosed {
		t.Errorf("initial state = %v, want CLOSED", state)
	}
	if failures != 0 {
		t.Errorf("initial failures = %d, want 0", failures)
	}
	if successes != 0 {
		t.Errorf("initial successes = %d, want 0", successes)
	}
}

// --- Testes de Allow ---
// Previne: Requests permitidos quando circuit está aberto

func TestCircuitBreaker_Allow(t *testing.T) {
	t.Run("allows when closed", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			MaxFailures:  3,
			ResetTimeout: time.Hour,
		})

		err := cb.Allow()
		if err != nil {
			t.Errorf("Allow() when closed should return nil, got %v", err)
		}
	})

	t.Run("blocks when open", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			MaxFailures:  1,
			ResetTimeout: time.Hour, // Long timeout to stay open
		})

		// Trigger opening
		cb.RecordFailure()

		err := cb.Allow()
		if err == nil {
			t.Error("Allow() when open should return error")
		}
	})

	t.Run("allows in half-open after timeout", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			MaxFailures:  1,
			ResetTimeout: time.Millisecond, // Very short timeout
		})

		// Open the circuit
		cb.RecordFailure()

		// Wait for reset timeout
		time.Sleep(10 * time.Millisecond)

		err := cb.Allow()
		if err != nil {
			t.Errorf("Allow() after timeout should return nil, got %v", err)
		}

		state, _, _ := cb.GetState()
		if state != CircuitHalfOpen {
			t.Errorf("state after timeout should be HALF-OPEN, got %v", state)
		}
	})
}

// --- Testes de RecordSuccess ---
// Previne: Circuit que não fecha após sucessos

func TestCircuitBreaker_RecordSuccess(t *testing.T) {
	t.Run("resets failure count when closed", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			MaxFailures: 5,
		})

		// Add some failures
		cb.RecordFailure()
		cb.RecordFailure()

		// Success should reset
		cb.RecordSuccess()

		_, failures, _ := cb.GetState()
		if failures != 0 {
			t.Errorf("failures after success = %d, want 0", failures)
		}
	})

	t.Run("closes circuit after enough successes in half-open", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			MaxFailures:      1,
			ResetTimeout:     time.Millisecond,
			SuccessThreshold: 2,
		})

		// Open and move to half-open
		cb.RecordFailure()
		time.Sleep(5 * time.Millisecond)
		_ = cb.Allow() // Moves to half-open

		// Record successes
		cb.RecordSuccess()
		state, _, _ := cb.GetState()
		if state != CircuitHalfOpen {
			t.Errorf("state after 1 success = %v, want HALF-OPEN", state)
		}

		cb.RecordSuccess()
		state, _, _ = cb.GetState()
		if state != CircuitClosed {
			t.Errorf("state after 2 successes = %v, want CLOSED", state)
		}
	})
}

// --- Testes de RecordFailure ---
// Previne: Circuit que não abre após falhas suficientes

func TestCircuitBreaker_RecordFailure(t *testing.T) {
	t.Run("opens after max failures", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			MaxFailures: 3,
		})

		cb.RecordFailure()
		cb.RecordFailure()

		state, _, _ := cb.GetState()
		if state != CircuitClosed {
			t.Errorf("state after 2 failures = %v, want CLOSED", state)
		}

		cb.RecordFailure()
		state, _, _ = cb.GetState()
		if state != CircuitOpen {
			t.Errorf("state after 3 failures = %v, want OPEN", state)
		}
	})

	t.Run("immediately opens from half-open on failure", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			MaxFailures:      1,
			ResetTimeout:     time.Millisecond,
			SuccessThreshold: 3,
		})

		// Open and move to half-open
		cb.RecordFailure()
		time.Sleep(5 * time.Millisecond)
		_ = cb.Allow()

		state, _, _ := cb.GetState()
		if state != CircuitHalfOpen {
			t.Fatalf("expected HALF-OPEN, got %v", state)
		}

		// Any failure should immediately open
		cb.RecordFailure()
		state, _, _ = cb.GetState()
		if state != CircuitOpen {
			t.Errorf("state after failure in half-open = %v, want OPEN", state)
		}
	})
}

// --- Testes de CircuitBreakerManager ---
// Previne: Gerenciamento incorreto de múltiplos providers

func TestCircuitBreakerManager(t *testing.T) {
	t.Run("allows unknown provider by default", func(t *testing.T) {
		mgr := NewCircuitBreakerManager()

		err := mgr.Allow("unknown-provider")
		if err != nil {
			t.Errorf("Allow() for unknown provider should return nil, got %v", err)
		}
	})

	t.Run("manages multiple providers independently", func(t *testing.T) {
		mgr := NewCircuitBreakerManager()

		// Usar timeout longo para garantir que o circuit fique aberto
		mgr.SetCircuitBreaker("provider-a", CircuitBreakerConfig{
			MaxFailures:  1,
			ResetTimeout: time.Hour, // Longo para não ir para HALF-OPEN
		})
		mgr.SetCircuitBreaker("provider-b", CircuitBreakerConfig{
			MaxFailures:  1,
			ResetTimeout: time.Hour,
		})

		// Open provider-a
		mgr.RecordFailure("provider-a")

		// provider-a should be blocked
		err := mgr.Allow("provider-a")
		if err == nil {
			t.Error("provider-a should be blocked after failure")
		}

		// provider-b should still work
		err = mgr.Allow("provider-b")
		if err != nil {
			t.Errorf("provider-b should still be allowed, got %v", err)
		}
	})

	t.Run("GetStatus returns false for unknown provider", func(t *testing.T) {
		mgr := NewCircuitBreakerManager()

		state, failures, successes, exists := mgr.GetStatus("unknown")
		if exists {
			t.Error("exists should be false for unknown provider")
		}
		if state != CircuitClosed {
			t.Errorf("state = %v, want CLOSED", state)
		}
		if failures != 0 || successes != 0 {
			t.Error("failures and successes should be 0 for unknown provider")
		}
	})

	t.Run("GetStatus returns correct values", func(t *testing.T) {
		mgr := NewCircuitBreakerManager()
		mgr.SetCircuitBreaker("test", CircuitBreakerConfig{MaxFailures: 5})

		mgr.RecordFailure("test")
		mgr.RecordFailure("test")

		state, failures, successes, exists := mgr.GetStatus("test")
		if !exists {
			t.Error("exists should be true")
		}
		if state != CircuitClosed {
			t.Errorf("state = %v, want CLOSED", state)
		}
		if failures != 2 {
			t.Errorf("failures = %d, want 2", failures)
		}
		if successes != 0 {
			t.Errorf("successes = %d, want 0", successes)
		}
	})
}

// --- Testes de concorrência ---
// Previne: Race conditions em uso concorrente

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		MaxFailures:      100,
		ResetTimeout:     time.Hour,
		SuccessThreshold: 10,
	})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(3)

		go func() {
			defer wg.Done()
			_ = cb.Allow()
		}()

		go func() {
			defer wg.Done()
			cb.RecordSuccess()
		}()

		go func() {
			defer wg.Done()
			cb.RecordFailure()
		}()
	}

	wg.Wait()
	// Se chegou aqui sem panic ou race, o teste passou
}

func TestCircuitBreakerManager_ConcurrentAccess(t *testing.T) {
	mgr := NewCircuitBreakerManager()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(4)

		go func(i int) {
			defer wg.Done()
			mgr.SetCircuitBreaker("provider", CircuitBreakerConfig{MaxFailures: 10})
		}(i)

		go func() {
			defer wg.Done()
			_ = mgr.Allow("provider")
		}()

		go func() {
			defer wg.Done()
			mgr.RecordSuccess("provider")
		}()

		go func() {
			defer wg.Done()
			mgr.RecordFailure("provider")
		}()
	}

	wg.Wait()
}

// --- Benchmark ---

func BenchmarkCircuitBreaker_Allow(b *testing.B) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		MaxFailures:  100,
		ResetTimeout: time.Hour,
	})

	for i := 0; i < b.N; i++ {
		_ = cb.Allow()
	}
}

func BenchmarkCircuitBreakerManager_Allow(b *testing.B) {
	mgr := NewCircuitBreakerManager()
	mgr.SetCircuitBreaker("test", CircuitBreakerConfig{MaxFailures: 100})

	for i := 0; i < b.N; i++ {
		_ = mgr.Allow("test")
	}
}

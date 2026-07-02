package pricing

import (
	"math"
	"testing"
)

// TestCost comprueba el coste local frente a un cálculo de referencia (SC-001, ±1%).
func TestCost(t *testing.T) {
	// claude-opus-4-6: Input 15, Output 75, CacheWrite 18.75, CacheRead 1.5 (USD/millón).
	cost, ok := Cost("claude-opus-4-6", 1_000_000, 1_000_000, 1_000_000, 1_000_000)
	if !ok {
		t.Fatalf("modelo conocido debe devolver ok=true")
	}
	want := 15.0 + 75.0 + 18.75 + 1.5
	if math.Abs(cost-want)/want > 0.01 {
		t.Errorf("coste fuera de ±1%%: got %v want %v", cost, want)
	}
}

// TestCost_UnknownModel: un modelo ausente de la tabla no bloquea; ok=false, coste 0.
func TestCost_UnknownModel(t *testing.T) {
	cost, ok := Cost("modelo-inexistente", 1000, 1000, 0, 0)
	if ok {
		t.Errorf("modelo desconocido debe devolver ok=false")
	}
	if cost != 0 {
		t.Errorf("modelo desconocido debe devolver coste 0, got %v", cost)
	}
}

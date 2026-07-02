package config

import (
	"testing"
)

func TestLoadOrCreateSalt_StableAndSecret(t *testing.T) {
	dir := t.TempDir()
	s1, err := LoadOrCreateSalt(dir)
	if err != nil {
		t.Fatalf("primera creación: %v", err)
	}
	if len(s1) != 64 { // 32 bytes en hex
		t.Errorf("salt debe ser 32 bytes hex (64 chars), got %d", len(s1))
	}
	s2, err := LoadOrCreateSalt(dir)
	if err != nil {
		t.Fatalf("segunda carga: %v", err)
	}
	if s1 != s2 {
		t.Errorf("el salt debe ser estable entre cargas: %q != %q", s1, s2)
	}
}

func TestLoadOrCreateMachineID_Stable(t *testing.T) {
	dir := t.TempDir()
	m1, err := LoadOrCreateMachineID(dir)
	if err != nil {
		t.Fatal(err)
	}
	m2, err := LoadOrCreateMachineID(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m1 == "" || m1 != m2 {
		t.Errorf("machine_id debe ser no vacío y estable: %q vs %q", m1, m2)
	}
}

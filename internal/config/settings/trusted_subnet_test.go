package settings

import "testing"

func TestTrustedSubnetDefaultEmpty(t *testing.T) {
	setupSettings(t)

	s, err := InitSettings()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.TrustedSubnet != "" {
		t.Errorf("expected empty trusted subnet, got %q", s.TrustedSubnet)
	}
}

func TestTrustedSubnetFromEnv(t *testing.T) {
	setupSettings(t)
	t.Setenv("TRUSTED_SUBNET", "192.168.1.0/24")

	s, err := InitSettings()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(s.TrustedSubnet) != "192.168.1.0/24" {
		t.Errorf("trusted subnet not propagated: got %q", s.TrustedSubnet)
	}
}

func TestTrustedSubnetFromFlag(t *testing.T) {
	setupSettings(t, "-t", "10.0.0.0/8")

	s, err := InitSettings()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(s.TrustedSubnet) != "10.0.0.0/8" {
		t.Errorf("trusted subnet not propagated: got %q", s.TrustedSubnet)
	}
}

func TestTrustedSubnetInvalidIsError(t *testing.T) {
	setupSettings(t)
	t.Setenv("TRUSTED_SUBNET", "not-a-cidr")

	if _, err := InitSettings(); err == nil {
		t.Fatal("expected error for invalid trusted subnet")
	}
}

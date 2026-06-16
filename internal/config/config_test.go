package config

import (
	"path/filepath"
	"testing"
)

func TestLoadMissingFile(t *testing.T) {
	orig := Path
	t.Cleanup(func() { Path = orig })
	Path = filepath.Join(t.TempDir(), "missing.toml")
	if _, err := Load(); err == nil {
		t.Error("expected error for missing config file")
	}
}

func TestSpecialPrefixEmpty(t *testing.T) {
	if specialPrefix("") {
		t.Error(`specialPrefix("") = true, want false`)
	}
}

func loadFixture(t *testing.T, path string) *Config {
	orig := Path
	t.Cleanup(func() { Path = orig })
	Path = path
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	return cfg
}

func TestExpandEnvVars(t *testing.T) {
	t.Setenv("TEST_HOST", "example.com")
	t.Setenv("TEST_USER", "alice")
	t.Setenv("TEST_IDENTITY", "/keys/id_ed25519")
	t.Setenv("TEST_PORT", "2222")
	t.Setenv("TEST_LOCAL", "9000")
	t.Setenv("TEST_REMOTE", "localhost:8080")

	cfg := loadFixture(t, "../../test/testdata/config/expand/vars.toml")

	tun := cfg.Tunnels[0]
	if tun.Host != "example.com" {
		t.Errorf("Host = %q, want %q", tun.Host, "example.com")
	}
	if tun.User != "alice" {
		t.Errorf("User = %q, want %q", tun.User, "alice")
	}
	if tun.IdentityFile != "/keys/id_ed25519" {
		t.Errorf("IdentityFile = %q, want %q", tun.IdentityFile, "/keys/id_ed25519")
	}
	if tun.Port.String() != "2222" {
		t.Errorf("Port = %q, want %q", tun.Port.String(), "2222")
	}
	if tun.LocalAddress.String() != "9000" {
		t.Errorf("LocalAddress = %q, want %q", tun.LocalAddress.String(), "9000")
	}
	if tun.RemoteAddress.String() != "localhost:8080" {
		t.Errorf("RemoteAddress = %q, want %q", tun.RemoteAddress.String(), "localhost:8080")
	}
}

func TestExpandDefault(t *testing.T) {
	t.Setenv("TEST_SET", "real")
	t.Setenv("TEST_EMPTY", "")
	// TEST_UNSET is unset

	cfg := loadFixture(t, "../../test/testdata/config/expand/defaults.toml")

	cases := map[string]string{
		"set":   "real",
		"empty": "fallback",
		"unset": "fallback",
	}
	for name, want := range cases {
		if got := cfg.TunnelsMap[name].Host; got != want {
			t.Errorf("tunnel %q Host = %q, want %q", name, got, want)
		}
	}
}

func TestExpandUnsetIsEmpty(t *testing.T) {
	// A referenced-but-unset variable with no default expands to ""
	cfg := loadFixture(t, "../../test/testdata/config/expand/unset.toml")

	if got := cfg.Tunnels[0].Host; got != "" {
		t.Errorf("Host = %q, want empty string", got)
	}
}

func TestExpandFieldsNotExpanded(t *testing.T) {
	// Name and Group are identifiers and are intentionally not expanded
	t.Setenv("TEST_VAR", "expanded")

	cfg := loadFixture(t, "../../test/testdata/config/expand/literal_fields.toml")

	tun := cfg.Tunnels[0]
	if tun.Name != "tun_$TEST_VAR" {
		t.Errorf("Name = %q, want it left literal", tun.Name)
	}
	if tun.Group != "grp_$TEST_VAR" {
		t.Errorf("Group = %q, want it left literal", tun.Group)
	}
}

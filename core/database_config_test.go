package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDatabaseConfigValidation(t *testing.T) {
	for _, o := range []DatabaseOptions{{Username: "health", PasswordEnv: "DB_PASSWORD"}, {TLS: "required"}, {TLS: "disabled"}} {
		c := Config{Version: "v1", Targets: []Target{{Name: "db", Type: "database", Endpoint: "redis://localhost", Database: &o}}}
		if err := c.Validate(); err != nil {
			t.Fatal(err)
		}
		c.Targets[0].Type = "http"
		if err := c.Validate(); err == nil {
			t.Fatal("database block accepted on HTTP")
		}
	}
	for _, o := range []DatabaseOptions{{TLS: "skip-verify"}, {PasswordEnv: "bad=reference"}} {
		if err := (Config{Version: "v1", Targets: []Target{{Name: "db", Type: "database", Endpoint: "redis://localhost", Database: &o}}}).Validate(); err == nil {
			t.Fatal("unsafe options accepted")
		}
	}
}

func TestDatabaseFixtureParity(t *testing.T) {
	for _, kind := range []string{"valid", "invalid"} {
		paths, err := filepath.Glob("../tests/spec/fixtures/" + kind + "/configuration.database*.json")
		if err != nil || len(paths) == 0 {
			t.Fatal(err)
		}
		for _, path := range paths {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			_, err = LoadConfig(strings.NewReader(string(data)))
			if (err == nil) != (kind == "valid") {
				t.Errorf("fixture %s: %v", path, err)
			}
		}
	}
}

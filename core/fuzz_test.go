package core

import (
	"bytes"
	"testing"
)

func FuzzLoadConfig(f *testing.F) {
	f.Add([]byte("version: v1\ntargets:\n- name: api\n  type: http\n  endpoint: https://example.com\n"))
	f.Add([]byte("version: v1\ntargets: []\n"))
	f.Add([]byte("x: &x [*x]"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 65536 {
			t.Skip()
		}
		config, err := LoadConfig(bytes.NewReader(data))
		if err == nil && config.Validate() != nil {
			t.Fatal("loader accepted invalid config")
		}
	})
}

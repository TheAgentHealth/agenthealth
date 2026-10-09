package a2a

import "testing"

func FuzzA2ACard(f *testing.F) {
	f.Add([]byte(`{"name":"agent","supportedInterfaces":[{"url":"https://example.com/rpc","protocolBinding":"JSONRPC","protocolVersion":"1.0"}]}`))
	f.Add([]byte(`{"securityRequirements":[{"schemes":null}]}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 65536 {
			t.Skip()
		}
		_, _ = v1Card(data)
	})
}

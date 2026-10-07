package mcp

import "encoding/json"

// Validate the content discriminant and its required fields without rendering
// or exposing tool output. Empty text and additional extension fields are valid.
func validContent(raw json.RawMessage) bool {
	var c map[string]json.RawMessage
	if json.Unmarshal(raw, &c) != nil || c == nil {
		return false
	}
	var kind string
	if json.Unmarshal(c["type"], &kind) != nil {
		return false
	}
	switch kind {
	case "text":
		return stringField(c, "text")
	case "image", "audio":
		return stringField(c, "data") && stringField(c, "mimeType")
	case "resource_link":
		return stringField(c, "uri") && stringField(c, "name")
	case "resource":
		var r map[string]json.RawMessage
		if json.Unmarshal(c["resource"], &r) != nil || r == nil || !stringField(r, "uri") {
			return false
		}
		if mime, ok := r["mimeType"]; ok && !stringField(map[string]json.RawMessage{"mimeType": mime}, "mimeType") {
			return false
		}
		return stringField(r, "text") || stringField(r, "blob")
	default:
		return false
	}
}

func stringField(fields map[string]json.RawMessage, name string) bool {
	raw, ok := fields[name]
	var value string
	return ok && string(raw) != "null" && json.Unmarshal(raw, &value) == nil
}

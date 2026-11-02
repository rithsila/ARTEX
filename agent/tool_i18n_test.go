package agent

import (
	"encoding/json"
	"testing"
)

// TestBuiltinToolSeedsAreEnglish guards the English tool catalog: no built-in
// seed description or parameter help may still contain Chinese text.
func TestBuiltinToolSeedsAreEnglish(t *testing.T) {
	for _, s := range BuiltinToolSeeds() {
		if HasCJK(s.Desc) {
			t.Errorf("tool %s: description still has Chinese: %q", s.Key, s.Desc)
		}
		b, err := json.Marshal(s.Schema)
		if err != nil {
			t.Errorf("tool %s: schema marshal: %v", s.Key, err)
			continue
		}
		var obj map[string]any
		if err := json.Unmarshal(b, &obj); err != nil {
			t.Errorf("tool %s: schema unmarshal: %v", s.Key, err)
			continue
		}
		if SchemaHasCJK(obj) {
			t.Errorf("tool %s: schema still has Chinese text", s.Key)
		}
	}
}

// TestBuiltinToolSeedsKeepSchemaShape ensures localization never drops the
// structure a tool handler depends on.
func TestBuiltinToolSeedsKeepSchemaShape(t *testing.T) {
	for _, s := range BuiltinToolSeeds() {
		if len(s.Schema) == 0 {
			t.Errorf("tool %s: schema became empty", s.Key)
		}
	}
}

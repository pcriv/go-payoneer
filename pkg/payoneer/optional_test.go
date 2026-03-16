package payoneer_test

import (
	"encoding/json"
	"testing"

	"github.com/pcriv/go-payoneer/pkg/payoneer"
)

func TestOptional(t *testing.T) {
	t.Run("Some", func(t *testing.T) {
		opt := payoneer.Some("value")
		if val, ok := opt.Get(); !ok || val != "value" {
			t.Errorf("expected value 'value' and ok=true, got %v and %v", val, ok)
		}
	})

	t.Run("None", func(t *testing.T) {
		opt := payoneer.None[string]()
		if val, ok := opt.Get(); ok {
			t.Errorf("expected ok=false, got %v and ok=true", val)
		}
	})

	t.Run("OrDefault", func(t *testing.T) {
		optSome := payoneer.Some("value")
		if val := optSome.OrDefault("default"); val != "value" {
			t.Errorf("expected 'value', got %v", val)
		}

		optNone := payoneer.None[string]()
		if val := optNone.OrDefault("default"); val != "default" {
			t.Errorf("expected 'default', got %v", val)
		}
	})

	t.Run("JSON Marshaling", func(t *testing.T) {
		type testStruct struct {
			Field payoneer.Optional[string] `json:"field"`
		}

		sSome := testStruct{Field: payoneer.Some("value")}
		dataSome, err := json.Marshal(sSome)
		if err != nil {
			t.Fatalf("failed to marshal Some: %v", err)
		}
		if string(dataSome) != `{"field":"value"}` {
			t.Errorf("expected `{\"field\":\"value\"}`, got %s", string(dataSome))
		}

		sNone := testStruct{Field: payoneer.None[string]()}
		dataNone, err := json.Marshal(sNone)
		if err != nil {
			t.Fatalf("failed to marshal None: %v", err)
		}
		if string(dataNone) != `{"field":null}` {
			t.Errorf("expected `{\"field\":null}`, got %s", string(dataNone))
		}
	})

	t.Run("JSON Unmarshaling", func(t *testing.T) {
		type testStruct struct {
			Field payoneer.Optional[string] `json:"field"`
		}

		var sSome testStruct
		err := json.Unmarshal([]byte(`{"field":"value"}`), &sSome)
		if err != nil {
			t.Fatalf("failed to unmarshal value: %v", err)
		}
		if val, ok := sSome.Field.Get(); !ok || val != "value" {
			t.Errorf("expected value 'value' and ok=true, got %v and %v", val, ok)
		}

		var sNull testStruct
		err = json.Unmarshal([]byte(`{"field":null}`), &sNull)
		if err != nil {
			t.Fatalf("failed to unmarshal null: %v", err)
		}
		if _, ok := sNull.Field.Get(); ok {
			t.Errorf("expected ok=false for null field, got ok=true")
		}

		var sMissing testStruct
		err = json.Unmarshal([]byte(`{}`), &sMissing)
		if err != nil {
			t.Fatalf("failed to unmarshal empty: %v", err)
		}
		if _, ok := sMissing.Field.Get(); ok {
			t.Errorf("expected ok=false for missing field, got ok=true")
		}
	})
}

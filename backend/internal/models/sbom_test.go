package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONB_Value(t *testing.T) {
	tests := []struct {
		name    string
		jsonb   JSONB
		wantErr bool
	}{
		{
			name: "valid JSONB",
			jsonb: JSONB{
				"key1": "value1",
				"key2": 123,
				"key3": true,
			},
			wantErr: false,
		},
		{
			name:    "nil JSONB",
			jsonb:   nil,
			wantErr: false,
		},
		{
			name:    "empty JSONB",
			jsonb:   JSONB{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.jsonb.Value()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.jsonb == nil {
					assert.Nil(t, val)
				} else {
					assert.NotNil(t, val)
					// Verify we can unmarshal back
					var result map[string]interface{}
					err = json.Unmarshal(val.([]byte), &result)
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestJSONB_Scan(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
		wantNil bool
	}{
		{
			name:    "valid JSON bytes",
			input:   []byte(`{"key":"value","number":42}`),
			wantErr: false,
			wantNil: false,
		},
		{
			name:    "nil value",
			input:   nil,
			wantErr: false,
			wantNil: true,
		},
		{
			name:    "empty JSON object",
			input:   []byte(`{}`),
			wantErr: false,
			wantNil: false,
		},
		{
			name:    "invalid JSON",
			input:   []byte(`{invalid json}`),
			wantErr: true,
			wantNil: false,
		},
		{
			name:    "non-byte value",
			input:   "not bytes",
			wantErr: true,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var j JSONB
			err := j.Scan(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.wantNil {
					assert.Nil(t, j)
				} else {
					assert.NotNil(t, j)
				}
			}
		})
	}
}

func TestJSONB_RoundTrip(t *testing.T) {
	original := JSONB{
		"name":    "test",
		"count":   42,
		"enabled": true,
		"nested": map[string]interface{}{
			"key": "value",
		},
	}

	// Convert to driver value
	val, err := original.Value()
	require.NoError(t, err)

	// Scan back
	var scanned JSONB
	err = scanned.Scan(val)
	require.NoError(t, err)

	// Verify equality
	assert.Equal(t, original["name"], scanned["name"])
	assert.Equal(t, float64(42), scanned["count"]) // JSON unmarshals numbers as float64
	assert.Equal(t, original["enabled"], scanned["enabled"])
}

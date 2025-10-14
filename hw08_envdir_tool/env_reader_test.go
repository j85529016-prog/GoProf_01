package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadDir(t *testing.T) {
	testdataDir := filepath.Join("testdata", "env")

	env, err := ReadDir(testdataDir)
	require.NoError(t, err, "ReadDir should not return error")
	require.Len(t, env, 5, "expected 5 environment variables")

	tests := []struct {
		name       string
		wantValue  string
		wantRemove bool
	}{
		{"BAR", "bar", false},
		{"EMPTY", "", true},
		{"FOO", "   foo\nwith new line", false},
		{"HELLO", "\"hello\"", false},
		{"UNSET", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev, ok := env[tt.name]
			require.True(t, ok, "variable %s not found", tt.name)
			require.Equal(t, tt.wantValue, ev.Value, "Value mismatch for %s", tt.name)
			require.Equal(t, tt.wantRemove, ev.NeedRemove, "NeedRemove mismatch for %s", tt.name)
		})
	}
}

package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestValidOptions for cache configs
func TestValidOptions(t *testing.T) {
	testCases := []struct {
		name     string
		input    Options
		expected struct {
			ItemTTL    time.Duration
			GCInterval time.Duration
		}
	}{
		{
			name: "valid values",
			input: Options{
				ItemTTL:    2 * time.Minute,
				GCInterval: 30 * time.Second,
			},
			expected: struct {
				ItemTTL    time.Duration
				GCInterval time.Duration
			}{
				ItemTTL:    2 * time.Minute,
				GCInterval: 30 * time.Second,
			},
		},
		{
			name:  "zero values should backfill",
			input: Options{},
			expected: struct {
				ItemTTL    time.Duration
				GCInterval time.Duration
			}{
				ItemTTL:    DefaultItemTTL,
				GCInterval: DefaultGCInterval,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(it *testing.T) {
			var (
				validate = assert.New(it)

				err = tc.input.Validate() // run business logic
			)
			validate.NoError(err, "unexpected error from valid input: %+v", tc.input)

			// check values that can be defaulted
			validate.Equal(tc.expected.ItemTTL, tc.input.ItemTTL, "ItemTTL mismatch")
			validate.Equal(tc.expected.GCInterval, tc.input.GCInterval, "GCInterval mismatch")
		})
	}
}

// TestBadOptions should error out on invalid configs
func TestBadOptions(t *testing.T) {
	var (
		testCases = []struct {
			name  string
			input Options
		}{
			{
				name: "negative ItemTTL",
				input: Options{
					ItemTTL:    -1 * time.Second,
					GCInterval: 10 * time.Second,
				},
			},
			{
				name: "negative GCInterval",
				input: Options{
					ItemTTL:    10 * time.Second,
					GCInterval: -5 * time.Second,
				},
			},
		}
	)
	for _, tc := range testCases {
		t.Run(tc.name, func(it *testing.T) {
			var (
				validate = assert.New(it)

				err = tc.input.Validate() // run business logic
			)
			validate.Error(err, "expected error from invalid input: `%+v`", tc.input)
		})
	}
}

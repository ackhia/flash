package node

import (
	"bytes"
	"testing"

	"github.com/ackhia/flash/models"
)

func TestMergeTxs(t *testing.T) {
	n := Node{}

	tests := []struct {
		name     string
		tx1      map[string][]models.Tx
		tx2      map[string][]models.Tx
		expected map[string][]models.Tx
	}{
		{
			name: "Basic Merge",
			tx1: map[string][]models.Tx{
				"key1": {{Sig: []byte("sig1")}, {Sig: []byte("sig2")}},
			},
			tx2: map[string][]models.Tx{
				"key2": {{Sig: []byte("sig3")}, {Sig: []byte("sig4")}},
			},
			expected: map[string][]models.Tx{
				"key1": {{Sig: []byte("sig1")}, {Sig: []byte("sig2")}},
				"key2": {{Sig: []byte("sig3")}, {Sig: []byte("sig4")}},
			},
		},
		{
			name: "Duplicate Transactions",
			tx1: map[string][]models.Tx{
				"key1": {{Sig: []byte("sig1")}},
			},
			tx2: map[string][]models.Tx{
				"key1": {{Sig: []byte("sig1")}, {Sig: []byte("sig2")}},
			},
			expected: map[string][]models.Tx{
				"key1": {{Sig: []byte("sig1")}, {Sig: []byte("sig2")}},
			},
		},
		{
			name:     "Empty Inputs",
			tx1:      map[string][]models.Tx{},
			tx2:      map[string][]models.Tx{},
			expected: map[string][]models.Tx{},
		},
		{
			name: "Partial Overlap",
			tx1: map[string][]models.Tx{
				"key1": {{Sig: []byte("sig1")}},
			},
			tx2: map[string][]models.Tx{
				"key1": {{Sig: []byte("sig2")}},
				"key2": {{Sig: []byte("sig3")}},
			},
			expected: map[string][]models.Tx{
				"key1": {{Sig: []byte("sig1")}, {Sig: []byte("sig2")}},
				"key2": {{Sig: []byte("sig3")}},
			},
		},
		{
			name: "No Overlap",
			tx1: map[string][]models.Tx{
				"key1": {{Sig: []byte("sig1")}},
			},
			tx2: map[string][]models.Tx{
				"key2": {{Sig: []byte("sig2")}},
			},
			expected: map[string][]models.Tx{
				"key1": {{Sig: []byte("sig1")}},
				"key2": {{Sig: []byte("sig2")}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := n.mergeTxs(tt.tx1, tt.tx2)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
			}

			for key, expectedTxs := range tt.expected {
				resultTxs, exists := result[key]
				if !exists {
					t.Errorf("Key %s missing in result", key)
					continue
				}
				if len(resultTxs) != len(expectedTxs) {
					t.Errorf("For key %s, expected %d transactions, got %d", key, len(expectedTxs), len(resultTxs))
					continue
				}
				for i, expectedTx := range expectedTxs {
					if !bytes.Equal(resultTxs[i].Sig, expectedTx.Sig) {
						t.Errorf("For key %s, expected Sig %s at index %d, got %s", key, expectedTx.Sig, i, resultTxs[i].Sig)
					}
				}
			}
		})
	}
}

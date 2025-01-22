package node

import (
	"bytes"
	"testing"

	"github.com/ackhia/flash/models"
	"github.com/stretchr/testify/assert"
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

func TestCalcBalances_ValidTransactions(t *testing.T) {
	node := &Node{
		genesis: map[string]float64{"Alice": 100.0, "Bob": 50.0},
		Txs: map[string][]models.Tx{
			"Tx1": {
				{SequenceNum: 0, From: "Alice", To: "Bob", Amount: 30.0},
				{SequenceNum: 1, From: "Bob", To: "Alice", Amount: 20.0},
			},
		},
		Balances: map[string]float64{},
	}

	err := node.calcBalances()
	assert.NoError(t, err, "calcBalances should not return an error")

	expectedBalances := map[string]float64{
		"Alice": 90.0,
		"Bob":   60.0,
	}
	assert.Equal(t, expectedBalances, node.Balances, "balances should match expected values")
}

func TestCalcBalances_InvalidSequence(t *testing.T) {
	node := &Node{
		genesis: map[string]float64{"Alice": 100.0},
		Txs: map[string][]models.Tx{
			"Tx1": {
				{SequenceNum: 0, From: "Alice", To: "Bob", Amount: 30.0},
				{SequenceNum: 2, From: "Bob", To: "Alice", Amount: 20.0}, // Invalid sequence
			},
		},
		Balances: map[string]float64{},
	}

	err := node.calcBalances()
	assert.Error(t, err, "calcBalances should return an error for invalid sequence")
	assert.Equal(t, "transactions must be ordered by sequence number", err.Error())
}

func TestCalcBalances_NegativeBalances(t *testing.T) {
	node := &Node{
		genesis: map[string]float64{"Alice": 100.0},
		Txs: map[string][]models.Tx{
			"Tx1": {
				{SequenceNum: 0, From: "Alice", To: "Bob", Amount: 120.0}, // Alice would have a negative balance
			},
		},
		Balances: map[string]float64{},
	}

	err := node.calcBalances()
	assert.Error(t, err, "calcBalances should return an error for negative balances")
	assert.Equal(t, "negative balances not allowed", err.Error())
}

func TestCalcBalances_EmptyGenesisAndTransactions(t *testing.T) {
	node := &Node{
		genesis:  map[string]float64{},
		Txs:      map[string][]models.Tx{},
		Balances: map[string]float64{},
	}

	err := node.calcBalances()
	assert.NoError(t, err, "calcBalances should not return an error for empty genesis and transactions")

	expectedBalances := map[string]float64{}
	assert.Equal(t, expectedBalances, node.Balances, "balances should be empty when genesis and transactions are empty")
}

func TestCalcBalances_NewAccountsFromTransactions(t *testing.T) {
	node := &Node{
		genesis: map[string]float64{"Alice": 50.0},
		Txs: map[string][]models.Tx{
			"Tx1": {
				{SequenceNum: 0, From: "Alice", To: "Charlie", Amount: 20.0}, // "Charlie" is a new account
			},
		},
		Balances: map[string]float64{},
	}

	err := node.calcBalances()
	assert.NoError(t, err, "calcBalances should not return an error for new accounts introduced by transactions")

	expectedBalances := map[string]float64{
		"Alice":   30.0,
		"Charlie": 20.0,
	}
	assert.Equal(t, expectedBalances, node.Balances, "balances should include new accounts from transactions")
}

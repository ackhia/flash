package node

import (
	"fmt"

	"github.com/ackhia/flash/models"
)

func (n Node) mergeTxs(tx1, tx2 map[string][]models.Tx) map[string][]models.Tx {
	superSet := make(map[string][]models.Tx)

	for key, tx := range tx1 {
		superSet[key] = append(superSet[key], tx...)
	}

	for key, txs := range tx2 {
		seen := make(map[string]struct{})
		for _, tx := range superSet[key] {
			seen[string(tx.Sig)] = struct{}{}
		}

		for _, tx := range txs {
			if _, exists := seen[string(tx.Sig)]; !exists {
				superSet[key] = append(superSet[key], tx)
			}
		}
	}

	return superSet
}

func (n *Node) calcBalances() error {
	for p, b := range n.genesis {
		n.Balances[p] = b
	}

	for _, txs := range n.Txs {
		for i := 0; i < len(txs); i++ {
			if txs[i].SequenceNum != i {
				return fmt.Errorf("transactions must be ordered by sequence number")
			}

			if _, ok := n.Balances[txs[i].From]; !ok {
				n.Balances[txs[i].From] = 0
			}

			if _, ok := n.Balances[txs[i].To]; !ok {
				n.Balances[txs[i].To] = 0
			}

			n.Balances[txs[i].From] -= txs[i].Amount
			n.Balances[txs[i].To] += txs[i].Amount
			if n.Balances[txs[i].From] < 0 {
				return fmt.Errorf("negative balances not allowed")
			}
		}
	}

	return nil
}

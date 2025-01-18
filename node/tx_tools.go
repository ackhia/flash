package node

import "github.com/ackhia/flash/models"

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

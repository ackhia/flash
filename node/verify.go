package node

import (
	"fmt"

	"github.com/ackhia/flash/models"
)

func (n Node) isVerifierConsensus(tx *models.Tx) (bool, error) {

	var verifierTotalBalance float64
	for _, v := range tx.Verifiers {
		b, ok := n.Balances[v.ID]
		if ok {
			verifierTotalBalance += b
		}
	}

	if verifierTotalBalance <= n.TotalCoins/2 {
		return false, fmt.Errorf("verifier total balances was less thsn 50%% of available coins")
	}
	return true, nil
}

package models

type Verifier struct {
	ID  string `json:"id"`
	Sig []byte `json:"sig"`
}

type Tx struct {
	SequenceNum int        `json:"sequenceNum"`
	From        string     `json:"from"`
	To          string     `json:"to"`
	Pubkey      []byte     `json:"pubKey"`
	Amount      float64    `json:"amount"`
	Sig         []byte     `json:"sig"`
	Verifiers   []Verifier `json:"verifiers"`
	Comitted    bool       `json:"-"`
}

package models

type Verifier struct {
	ID  string `json:"id"`
	Sig []byte `json:"sig"`
}

type Tx struct {
	SequenceNum int        `json:"sequenceNum"`
	From        string     `json:"from"`
	To          string     `json:"to"`
	Amount      float64    `json:"amount"`
	Sig         []byte     `json:"sig"`
	Verifiers   []Verifier `json:"verifiers"`
}

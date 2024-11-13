package models

type InstantSwap struct {
	ID            string
	QuotationID   string
	FromWalletID  string
	ToWalletID    string
	QuotationRate float64
	ExecutionRate float64
	SwapTxID0     string
	SwapTxID1     string
	QuoteTxID0    string
	QuoteTxID1    string
}

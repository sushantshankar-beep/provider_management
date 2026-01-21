package utils

type PayoutCalculation struct {
	BaseAmount  float64
	Commission  float64
	GST         float64
	NetPayable  float64
}

func CalculatePayout(amount, commissionPct, gstPct float64) PayoutCalculation {
	commission := amount * commissionPct / 100
	afterCommission := amount - commission
	gst := afterCommission * gstPct / 100
	netPayable := afterCommission - gst

	return PayoutCalculation{
		BaseAmount: RoundTo2(amount),
		Commission: RoundTo2(commission),
		GST:        RoundTo2(gst),
		NetPayable: RoundTo2(netPayable),
	}
}



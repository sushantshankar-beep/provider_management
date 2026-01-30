package utils

type PayoutCalculation struct {
	BaseAmount float64
	Commission float64
	GST        float64
	TDS        float64
	NetPayable float64
}

func CalculatePayout(amount, commissionPct, gstPct float64, tdsPercent float64) PayoutCalculation {
	commission := amount * commissionPct / 100
	afterCommission := amount - commission

	gst := 0.0
	tds := 0.0

	if tdsPercent > 0 {
		tds = afterCommission * tdsPercent / 100
	} else {
		gst = afterCommission * gstPct / 100
	}

	netPayable := afterCommission - gst - tds

	return PayoutCalculation{
		BaseAmount: RoundTo2(amount),
		Commission: RoundTo2(commission),
		GST:        RoundTo2(gst),
		TDS:        RoundTo2(tds),
		NetPayable: RoundTo2(netPayable),
	}
}

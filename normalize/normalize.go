package normalize

import (
	"encoding/json"
	"os"
	"runtime"
	"slices"

	"rinha2026/model"
	"rinha2026/quantization"
)

func LoadParams(path string) (model.Params, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return model.Params{}, err
	}
	var p model.Params
	if err := json.Unmarshal(data, &p); err != nil {
		return model.Params{}, err
	}

	data = nil
	runtime.GC()
	return p, nil
}

func LoadMccRisk(path string) (map[string]float64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m map[string]float64
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	data = nil
	runtime.GC()
	return m, nil
}

func clamp(v, max float64) float64 {
	if v > max {
		return max
	}
	if v < 0 {
		return 0
	}
	return v / max
}

func ToVectorQuantized(req model.FraudScoreRequest, p model.Params, mccRisk map[string]float64) [14]int16 {
	var result [14]int16

	result[0] = quantizeField(clamp(req.Transaction.Amount, p.MaxAmount))
	result[1] = quantizeField(clamp(req.Transaction.Installments, p.MaxInstallments))

	if req.Customer.AverageAmount > 0 {
		result[2] = quantizeField(clamp(req.Transaction.Amount/req.Customer.AverageAmount, p.AmountVsAvgRatio))
	} else {
		result[2] = 0
	}

	weekday := float64(req.Transaction.RequestedAt.Weekday())
	result[3] = quantizeField(clamp(weekday, 6))

	hour := float64(req.Transaction.RequestedAt.Hour())
	result[4] = quantizeField(clamp(hour, 23))

	if req.LastTransaction != nil {
		minutes := req.Transaction.RequestedAt.Sub(req.LastTransaction.Timestamp).Minutes()
		result[5] = quantizeField(clamp(minutes, p.MaxMinutes))
		result[6] = quantizeField(clamp(req.LastTransaction.KmFromCurrent, p.MaxKm))
	} else {
		result[5] = -1
		result[6] = -1
	}

	result[7] = quantizeField(clamp(req.Terminal.KmFromHome, p.MaxKm))

	result[8] = quantizeField(clamp(req.Customer.TxCount24H, p.MaxTxCount24h))

	result[9] = quantizeBool(req.Terminal.IsOnline)
	result[10] = quantizeBool(req.Terminal.CardPresent)

	isUnknownMerchant := !slices.Contains(req.Customer.KnownMerchants, req.Merchant.ID)
	result[11] = quantizeBool(isUnknownMerchant)

	mccRiskValue := 0.5
	if risk, exists := mccRisk[req.Merchant.Mcc]; exists {
		mccRiskValue = risk
	}
	result[12] = quantizeField(mccRiskValue)

	result[13] = quantizeField(clamp(req.Merchant.AverageAmount, p.MaxMerchantAvgAmount))

	return result
}

func quantizeField(v float64) int16 {
	return quantization.Quantize(v)
}

func quantizeBool(b bool) int16 {
	if b {
		return quantization.Quantize(1.0)
	}
	return quantization.Quantize(0.0)
}

package model

import "time"

type AccountBalance struct {
	AccountID string
	Balance   float64
	DateTime  time.Time
}

func NewAccountBalance(accountId string, balance float64, dateTime time.Time) *AccountBalance {
	return &AccountBalance{AccountID: accountId, Balance: balance, DateTime: dateTime}
}

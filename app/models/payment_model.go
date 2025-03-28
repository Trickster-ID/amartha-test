package models

import "time"

type Payment struct {
	WeekNumber int
	DueDate    *time.Time
	Amount     float64
	Paid       bool
}

type PaymentRecord struct {
	WeekNumber  int
	Amount      float64
	PaymentDate *time.Time
	Successful  bool
}

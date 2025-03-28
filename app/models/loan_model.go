package models

import "time"

type Loan struct {
	ID             string
	LoanAmount     float64
	InterestRate   float64
	TermInWeeks    int
	WeeklyPayment  float64
	StartDate      *time.Time
	Payments       []*Payment
	PaymentHistory []*PaymentRecord
}

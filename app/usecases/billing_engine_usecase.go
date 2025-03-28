package usecases

import (
	"amartha-test/app/models"
	"errors"
	"fmt"
	"time"
)

type IBillingEngineUseCase interface {
	CreateLoan(id string, principal float64, interestRate float64, termWeeks int, startDate time.Time) (*models.Loan, error)
	GetOutstanding(loanID string) float64
	IsDelinquent(loanID string) bool
	MakePayment(loanID string, amount float64, paymentDate time.Time) error
	GetPaymentSchedule(loanID string) ([]*models.Payment, error)
}

type BillingEngine struct {
	Loans map[string]*models.Loan
}

func NewBillingEngine() *BillingEngine {
	return &BillingEngine{
		Loans: make(map[string]*models.Loan),
	}
}

func (u *BillingEngine) CreateLoan(id string, loanAmount float64, interestRate float64, termInWeeks int, startDate time.Time) (*models.Loan, error) {
	// Check required param
	if loanAmount <= 0 || interestRate <= 0 || termInWeeks <= 0 {
		return nil, errors.New("invalid loan parameters")
	}

	// Check if id is existed
	if u.Loans[id] != nil {
		return nil, errors.New("loan already exists")
	}

	totalInterest := loanAmount * interestRate
	totalRepayment := loanAmount + totalInterest
	weeklyPayment := totalRepayment / float64(termInWeeks)

	loan := &models.Loan{
		ID:            id,
		LoanAmount:    loanAmount,
		InterestRate:  interestRate,
		TermInWeeks:   termInWeeks,
		WeeklyPayment: weeklyPayment,
		StartDate:     &startDate,
	}

	// Initialize payment schedule
	for i := 1; i <= termInWeeks; i++ {
		dueDate := startDate.AddDate(0, 0, 7*i)
		loan.Payments = append(loan.Payments, &models.Payment{
			WeekNumber: i,
			DueDate:    &dueDate,
			Amount:     weeklyPayment,
			Paid:       false,
		})
	}

	// Simulate save data
	u.Loans[id] = loan

	return loan, nil
}

func (u *BillingEngine) GetOutstanding(loanID string, currentTime *time.Time) float64 {
	// Check existed loan by id
	loan, exists := u.Loans[loanID]
	if !exists {
		return 0
	}

	// Check existed payment billing
	outstanding := 0.0
	for _, payment := range loan.Payments {
		if payment.DueDate.Before(*currentTime) && !payment.Paid {
			outstanding += payment.Amount
		}
	}

	return outstanding
}

func (u *BillingEngine) IsDelinquent(loanID string, currentTime *time.Time) bool {
	if currentTime == nil {
		return false
	}
	// Simulate get data by loan ID
	loan, exists := u.Loans[loanID]
	if !exists {
		return false
	}

	// Check last two payments
	missedCount := 0
	for i := len(loan.Payments) - 1; i >= 0; i-- {
		if loan.Payments[i].DueDate.Before(*currentTime) && !loan.Payments[i].Paid {
			missedCount++
			if missedCount > 2 {
				return true
			}
		}
	}

	return false
}

func (u *BillingEngine) MakePayment(loanID string, amount float64, paymentDate time.Time) error {
	// Check existed loan by id
	loan, exists := u.Loans[loanID]
	if !exists {
		return errors.New("loan not found")
	}

	// Check if payment match weekly amount
	if amount != loan.WeeklyPayment {
		return fmt.Errorf("payment must be exactly %.2f", loan.WeeklyPayment)
	}

	// Find the earliest unpaid payment
	var paymentToMake *models.Payment
	for i := range loan.Payments {
		if !loan.Payments[i].Paid {
			paymentToMake = loan.Payments[i]
			break
		}
	}

	if paymentToMake == nil {
		return errors.New("loan is already fully paid")
	}

	// Mark payment as paid
	paymentToMake.Paid = true

	// Record payment history
	loan.PaymentHistory = append(loan.PaymentHistory, &models.PaymentRecord{
		WeekNumber:  paymentToMake.WeekNumber,
		Amount:      amount,
		PaymentDate: &paymentDate,
		Successful:  true,
	})

	return nil
}

func (u *BillingEngine) GetPaymentSchedule(loanID string) ([]*models.Payment, error) {
	loan, exists := u.Loans[loanID]
	if !exists {
		return nil, errors.New("loan not found")
	}

	return loan.Payments, nil
}

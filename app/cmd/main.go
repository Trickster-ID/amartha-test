package main

import (
	"amartha-test/app/usecases"
	"fmt"
	"time"
)

func main() {
	// Initiate the UseCase
	billing := usecases.NewBillingEngine()

	// Create a new loan: Rp 5,000,000 for 50 weeks at 10% interest
	startDate := time.Now()
	_, err := billing.CreateLoan("100", 5000000, 0.10, 50, startDate)
	if err != nil {
		fmt.Println("Error creating loan:", err)
		return
	}

	// Get payment schedule
	schedule, _ := billing.GetPaymentSchedule("100")
	fmt.Println("Payment schedule:")
	for _, payment := range schedule {
		fmt.Printf("Week %d: Rp %.2f (Due: %s)\n", payment.WeekNumber, payment.Amount, payment.DueDate.Format("2006-01-02"))
	}

	fmt.Printf("\ncurrent time [%v] \n", startDate.AddDate(0, 0, 14))
	// Check initial outstanding
	fmt.Printf("\nInitial outstanding: Rp %.2f\n", billing.GetOutstanding("100", &startDate))

	// Make some payments
	err = billing.MakePayment("100", 110000, startDate.AddDate(0, 0, 7))
	if err != nil {
		panic(err)
	}
	err = billing.MakePayment("100", 110000, startDate.AddDate(0, 0, 14))
	if err != nil {
		panic(err)
	}

	// Check outstanding after payments
	fmt.Printf("Outstanding after 2 payments: Rp %.2f\n", billing.GetOutstanding("100", &startDate))

	startDate = startDate.AddDate(0, 0, 14)
	// Check delinquency status
	fmt.Println("Is delinquent?", billing.IsDelinquent("100", &startDate))

	// Simulate missed payments
	fmt.Println("\nSimulating missed 3payments...")
	startDate = startDate.AddDate(0, 0, 7*3).Add(time.Hour)
	fmt.Printf("current time [%v] \n", startDate)

	// Check delinquency status after missing 3 payments
	fmt.Println("Is delinquent after missing 3 payments?", billing.IsDelinquent("100", &startDate))
	fmt.Printf("Outstanding after missing 3 payments: Rp %.2f\n", billing.GetOutstanding("100", &startDate))

	// Make a payment to clear some delinquency
	err = billing.MakePayment("100", 110000, startDate)
	if err != nil {
		panic(err)
	}
	fmt.Println("\nAfter making one payment:")
	fmt.Println("Is delinquent?", billing.IsDelinquent("100", &startDate))
	fmt.Printf("Outstanding: Rp %.2f\n", billing.GetOutstanding("100", &startDate))
}

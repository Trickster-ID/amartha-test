package usecases

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCreateLoan(t *testing.T) {
	t.Run("successful loan creation", func(t *testing.T) {
		be := NewBillingEngine()
		startDate := time.Now()
		loan, err := be.CreateLoan("100", 5000000, 0.10, 50, startDate)

		assert.NoError(t, err)
		assert.NotNil(t, loan)
		assert.Equal(t, "100", loan.ID)
		assert.Equal(t, 5000000.0, loan.LoanAmount)
		assert.Equal(t, 0.10, loan.InterestRate)
		assert.Equal(t, 50, loan.TermInWeeks)
		assert.Equal(t, 110000.0, loan.WeeklyPayment)
		assert.Equal(t, 50, len(loan.Payments))
	})

	t.Run("invalid loan parameters", func(t *testing.T) {
		be := NewBillingEngine()
		startDate := time.Now()

		tests := []struct {
			name      string
			principal float64
			rate      float64
			term      int
		}{
			{"zero principal", 0, 0.10, 50},
			{"negative principal", -1000, 0.10, 50},
			{"zero interest", 5000000, 0, 50},
			{"negative interest", 5000000, -0.10, 50},
			{"zero term", 5000000, 0.10, 0},
			{"negative term", 5000000, 0.10, -50},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := be.CreateLoan("100", tt.principal, tt.rate, tt.term, startDate)
				assert.Error(t, err)
			})
		}
	})

	t.Run("duplicate loan ID", func(t *testing.T) {
		be := NewBillingEngine()
		startDate := time.Now()
		_, err := be.CreateLoan("100", 5000000, 0.10, 50, startDate)
		assert.NoError(t, err)

		_, err = be.CreateLoan("100", 6000000, 0.15, 60, startDate)
		assert.Error(t, err)
	})
}

func TestGetOutstanding(t *testing.T) {
	t.Run("initial outstanding", func(t *testing.T) {
		be := NewBillingEngine()
		startDate := time.Now()
		_, err := be.CreateLoan("100", 5000000, 0.10, 50, startDate)
		assert.NoError(t, err)

		outstanding := be.GetOutstanding("100", &startDate)
		assert.Equal(t, 0.00, outstanding) // 5,000,000 + 10% interest
	})

	t.Run("after some payments", func(t *testing.T) {
		be := NewBillingEngine()
		startDate := time.Now()
		_, err := be.CreateLoan("100", 5000000, 0.10, 50, startDate)
		assert.NoError(t, err)

		// Make 3 payments
		be.MakePayment("100", 110000, startDate.AddDate(0, 0, 7))
		be.MakePayment("100", 110000, startDate.AddDate(0, 0, 14))
		be.MakePayment("100", 110000, startDate.AddDate(0, 0, 21))

		after3weeks := startDate.AddDate(0, 0, 7*3)
		outstanding := be.GetOutstanding("100", &after3weeks)
		assert.Equal(t, 0.00, outstanding)
	})

	t.Run("fully paid loan", func(t *testing.T) {
		be := NewBillingEngine()
		startDate := time.Now()
		_, err := be.CreateLoan("100", 5000000, 0.10, 50, startDate)
		assert.NoError(t, err)

		// Make all payments
		for i := 1; i <= 50; i++ {
			be.MakePayment("100", 110000, startDate.AddDate(0, 0, 7*i))
		}

		outstanding := be.GetOutstanding("100", &startDate)
		assert.Equal(t, 0.0, outstanding)
	})

	t.Run("non-existent loan", func(t *testing.T) {
		be := NewBillingEngine()
		startDate := time.Now()
		outstanding := be.GetOutstanding("999", &startDate)
		assert.Equal(t, 0.0, outstanding)
	})
}

func TestIsDelinquent(t *testing.T) {
	t.Run("not delinquent - no missed payments", func(t *testing.T) {
		be := NewBillingEngine()
		startDate := time.Now()
		_, err := be.CreateLoan("100", 5000000, 0.10, 50, startDate)
		assert.NoError(t, err)

		// Make first payment
		be.MakePayment("100", 110000, startDate.AddDate(0, 0, 7))

		assert.False(t, be.IsDelinquent("100", &startDate))
	})

	t.Run("not delinquent - 1 missed payment", func(t *testing.T) {
		be := NewBillingEngine()
		startDate := time.Now()
		_, err := be.CreateLoan("100", 5000000, 0.10, 50, startDate)
		assert.NoError(t, err)

		// First payment made, second missed
		be.MakePayment("100", 110000, startDate.AddDate(0, 0, 7))

		assert.False(t, be.IsDelinquent("100", &startDate))
	})

	t.Run("delinquent - 3 consecutive missed payments", func(t *testing.T) {
		be := NewBillingEngine()
		startDate := time.Now()
		_, err := be.CreateLoan("100", 5000000, 0.10, 50, startDate)
		assert.NoError(t, err)

		// Make first 3 payments, then miss next 2
		be.MakePayment("100", 110000, startDate.AddDate(0, 0, 7))
		be.MakePayment("100", 110000, startDate.AddDate(0, 0, 14))
		be.MakePayment("100", 110000, startDate.AddDate(0, 0, 21))

		after6weeks := startDate.AddDate(0, 0, 7*6)
		after6weeks = after6weeks.Add(time.Hour)
		// Skip payments for weeks 4 and 5
		//fmt.Println(after6weeks)
		//jsondata, _ := json.MarshalIndent(be.Loans["100"], "", "  ")
		//fmt.Println(string(jsondata))
		assert.True(t, be.IsDelinquent("100", &after6weeks))
	})

	t.Run("not delinquent after catching up", func(t *testing.T) {
		be := NewBillingEngine()
		startDate := time.Now()
		_, err := be.CreateLoan("100", 5000000, 0.10, 50, startDate)
		assert.NoError(t, err)

		// Miss first 2 payments
		// Then make 2 payments to catch up
		be.MakePayment("100", 110000, startDate.AddDate(0, 0, 21))
		be.MakePayment("100", 110000, startDate.AddDate(0, 0, 28))

		assert.False(t, be.IsDelinquent("100", &startDate))
	})

	t.Run("non-existent loan", func(t *testing.T) {
		be := NewBillingEngine()
		startDate := time.Now()
		assert.False(t, be.IsDelinquent("999", &startDate))
	})
}

func TestMakePayment(t *testing.T) {
	t.Run("successful payment", func(t *testing.T) {
		be := NewBillingEngine()
		startDate := time.Now()
		_, err := be.CreateLoan("100", 5000000, 0.10, 50, startDate)
		assert.NoError(t, err)

		err = be.MakePayment("100", 110000, startDate.AddDate(0, 0, 7))
		assert.NoError(t, err)

		outstanding := be.GetOutstanding("100", &startDate)
		assert.Equal(t, 0.00, outstanding)
	})

	t.Run("incorrect payment amount", func(t *testing.T) {
		be := NewBillingEngine()
		startDate := time.Now()
		_, err := be.CreateLoan("100", 5000000, 0.10, 50, startDate)
		assert.NoError(t, err)

		err = be.MakePayment("100", 100000, startDate.AddDate(0, 0, 7))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "payment must be exactly")
	})

	t.Run("payment on non-existent loan", func(t *testing.T) {
		be := NewBillingEngine()
		err := be.MakePayment("999", 110000, time.Now())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "loan not found")
	})

	t.Run("fully paid loan", func(t *testing.T) {
		be := NewBillingEngine()
		startDate := time.Now()
		_, err := be.CreateLoan("100", 5000000, 0.10, 50, startDate)
		assert.NoError(t, err)

		// Make all payments
		for i := 1; i <= 50; i++ {
			be.MakePayment("100", 110000, startDate.AddDate(0, 0, 7*i))
		}

		// Try to make one more payment
		err = be.MakePayment("100", 110000, startDate.AddDate(0, 0, 7*51))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already fully paid")
	})
}

func TestGetPaymentSchedule(t *testing.T) {
	t.Run("get schedule for existing loan", func(t *testing.T) {
		be := NewBillingEngine()
		startDate := time.Now()
		_, err := be.CreateLoan("100", 5000000, 0.10, 50, startDate)
		assert.NoError(t, err)

		schedule, err := be.GetPaymentSchedule("100")
		assert.NoError(t, err)
		assert.Equal(t, 50, len(schedule))

		// Check first payment
		assert.Equal(t, 1, schedule[0].WeekNumber)
		assert.Equal(t, 110000.0, schedule[0].Amount)
		assert.Equal(t, startDate.AddDate(0, 0, 7), *schedule[0].DueDate)
		assert.False(t, schedule[0].Paid)

		// Check last payment
		assert.Equal(t, 50, schedule[49].WeekNumber)
		assert.Equal(t, 110000.0, schedule[49].Amount)
		assert.Equal(t, startDate.AddDate(0, 0, 7*50), *schedule[49].DueDate)
		assert.False(t, schedule[49].Paid)
	})

	t.Run("get schedule after some payments", func(t *testing.T) {
		be := NewBillingEngine()
		startDate := time.Now()
		_, err := be.CreateLoan("100", 5000000, 0.10, 50, startDate)
		assert.NoError(t, err)

		// Make first 3 payments
		be.MakePayment("100", 110000, startDate.AddDate(0, 0, 7))
		be.MakePayment("100", 110000, startDate.AddDate(0, 0, 14))
		be.MakePayment("100", 110000, startDate.AddDate(0, 0, 21))

		schedule, err := be.GetPaymentSchedule("100")
		assert.NoError(t, err)

		// Check first 3 payments are marked as paid
		assert.True(t, schedule[0].Paid)
		assert.True(t, schedule[1].Paid)
		assert.True(t, schedule[2].Paid)
		assert.False(t, schedule[3].Paid)
	})

	t.Run("get schedule for non-existent loan", func(t *testing.T) {
		be := NewBillingEngine()
		_, err := be.GetPaymentSchedule("999")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "loan not found")
	})
}

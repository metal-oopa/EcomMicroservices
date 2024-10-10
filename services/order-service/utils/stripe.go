package utils

import (
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/paymentintent"
)

func InitializeStripe(apiKey string) {
	stripe.Key = apiKey
}

func CreatePaymentIntent(amount int64, currency, paymentMethodID string) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(amount),
		Currency:      stripe.String(currency),
		PaymentMethod: stripe.String(paymentMethodID),
		Confirm:       stripe.Bool(true),
	}

	return paymentintent.New(params)
}

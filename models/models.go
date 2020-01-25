package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Invoice struct {
	Id     primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Line   []Line            `json:"line" bson:"line,omitempty"`
	Client Client            `json:"client" bson:"client,omitempty"`
	TaxTotal float64			`json:"tax_total" bson:"tax_total"`
	DiscountTotal float64			`json:"discount_total" bson:"discount_total"`
	Subtotal float64			`json:"subtotal" bson:"subtotal"`
	Total float64				`json:"total" bson:"total"`
	Payments PayInvoice		`json:"payments" bson:"payments"`
	Balance float64			`json:"balance" bson:"balance"`
}

type Line struct {
	Product string `json:"product,omitempty" bson:"product,omitempty"`
	Quantity int `json:"quantity,omitempty" bson:"quantity,omitempty"`
	Price float64 `json:"price,omitempty" bson:"price,omitempty"`
	PriceSrc float64 `json:"price_src" bson:"price_src,omitempty"`
	TaxRate float64 `json:"tax_rate,omitempty" bson:"tax_rate,omitempty"`
	DiscountRate float64 `json:"discount_rate,omitempty" bson:"discount_rate,omitempty"`
	Currency string `json:"currency,omitempty" bson:"currency,omitempty"`
}

type Client struct {
	Id    string `json:"id,omitempty" bson:"id,omitempty"`
	Name   string `json:"name,omitempty" bson:"name,omitempty"`
}

// Payment
type PayInvoice struct {
	InvoiceId string `json:"invoice_id,omitempty" bson:"invoice_id,omitempty"`
	Amount float64 `json:"amount,omitempty" bson:"amount,omitempty"`
}
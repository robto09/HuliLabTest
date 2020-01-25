package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)
import "regexp"
import "HuliTest/helper"
import "HuliTest/models"
import "HuliTest/webservice"

func main() {
	//Init Router
	r := mux.NewRouter()

	// invoice routes
	r.HandleFunc("/api/invoices", getInvoices).Methods("GET")
	r.HandleFunc("/api/invoices", createInvoice).Methods("POST")
	r.HandleFunc("/api/invoices/{id}", getInvoice).Methods("GET")
	r.HandleFunc("/api/invoices/{id}", updateInvoice).Methods("PUT")
	r.HandleFunc("/api/invoices/{id}", deleteInvoice).Methods("DELETE")

	// Pay invoice
	r.HandleFunc("/api/invoices/pay", createInvoicePayment).Methods("POST")

	// set port address
	log.Fatal(http.ListenAndServe(":7000", r))

}

func getInvoices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var invoices []models.Invoice

	//Connection mongoDB with helper class
	collection := helper.ConnectDB().Collection("invoices")

	cur, err := collection.Find(context.TODO(), bson.M{})

	if err != nil {
		helper.GetError(err, w)
		return
	}

	// Close the cursor once finished
	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {

		// create a value into which the single document can be decoded
		var invoice models.Invoice

		err := cur.Decode(&invoice) // decode
		if err != nil {
			log.Fatal(err)
		}

		// add item to array
		invoices = append(invoices, invoice)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(invoices) // encode
}

func createInvoice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var invoice models.Invoice
	var re = regexp.MustCompile(`(?m)^[0-9]+(-[0-9]+)+(-[0-9]+)+$`)

	fmt.Println(json.NewDecoder(r.Body).Decode(&invoice))


	_ = json.NewDecoder(r.Body).Decode(&invoice)


	if len(re.FindAllString(invoice.Client.Id, -1)) < 1 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode("client id must have atleast 3 group of integers")
	} else {
		rate := webservice.ConvertDollarsColones()
		lnArr := []models.Line{}
		var taxTotal = 0.0
		var discountTotal = 0.0
		var subTotal = 0.0
		var balance = 0.0
		if s, err := strconv.ParseFloat(rate, 64); err == nil {
			fmt.Println(s) // 3.14159265
			for i := 0; i < len(invoice.Line); i++  {
				line := invoice.Line[i]
				taxTotal = taxTotal + line.TaxRate
				discountTotal = discountTotal + line.DiscountRate
				if line.Currency != "CRC" {
					line.PriceSrc = s * line.Price
					line.Currency = "CRC"
				} else {
					line.PriceSrc = line.Price
				}
				subTotal = subTotal + line.PriceSrc
				lnArr = append(lnArr, line)
			}

		}
		invoice.TaxTotal = taxTotal
		invoice.DiscountTotal = discountTotal
		invoice.Subtotal = subTotal
		invoice.Total = taxTotal + subTotal - discountTotal
		invoice.Balance = balance
		invoice.Line = lnArr
		// connect db
		collection := helper.ConnectDB().Collection("invoices")

		// insert model.
		result, err := collection.InsertOne(context.TODO(), invoice)

		fmt.Println(invoice.Line)
		if err != nil {
			helper.GetError(err, w)
			return
		}

		json.NewEncoder(w).Encode(result)
	}
}

func getInvoice(w http.ResponseWriter, r *http.Request) {
	// set header.
	w.Header().Set("Content-Type", "application/json")

	var invoice models.Invoice

	var params = mux.Vars(r)

	// string to primitive.ObjectID
	id, _ := primitive.ObjectIDFromHex(params["id"])

	collection := helper.ConnectDB().Collection("invoices")

	filter := bson.M{"_id": id}
	err := collection.FindOne(context.TODO(), filter).Decode(&invoice)

	if err != nil {
		helper.GetError(err, w)
		return
	}

	json.NewEncoder(w).Encode(invoice)
}

func updateInvoice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var params = mux.Vars(r)

	//Get id from parameters
	id, _ := primitive.ObjectIDFromHex(params["id"])

	var invoice models.Invoice

	collection := helper.ConnectDB().Collection("invoices")

	// Create filter
	filter := bson.M{"_id": id}

	// Read update model from body request
	_ = json.NewDecoder(r.Body).Decode(&invoice)

	// prepare update model.
	update := bson.D{
		{"$set", bson.D{
			{"line", invoice.Line},
			{"client", bson.D{
				{"id", invoice.Client.Id},
				{"name", invoice.Client.Name},
			}},
		}},
	}

	err := collection.FindOneAndUpdate(context.TODO(), filter, update).Decode(&invoice)

	if err != nil {
		helper.GetError(err, w)
		return
	}

	invoice.Id = id

	json.NewEncoder(w).Encode(invoice)
}

func deleteInvoice(w http.ResponseWriter, r *http.Request) {
	// Set header
	w.Header().Set("Content-Type", "application/json")

	// get params
	var params = mux.Vars(r)

	id, err := primitive.ObjectIDFromHex(params["id"])

	collection := helper.ConnectDB().Collection("invoices")

	filter := bson.M{"_id": id}

	deleteResult, err := collection.DeleteOne(context.TODO(), filter)

	if err != nil {
		helper.GetError(err, w)
		return
	}

	json.NewEncoder(w).Encode(deleteResult)
}


// Invoice Payment
func createInvoicePayment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var pay models.PayInvoice
	var pay2 models.PayInvoice
	fmt.Println(json.NewDecoder(r.Body).Decode(&pay))

	_ = json.NewDecoder(r.Body).Decode(&pay)

	// Find if already a payment
	var id = pay.InvoiceId

	// connect db
	collection := helper.ConnectDB().Collection("payment")

	// prepare filter.
	filter := bson.M{"invoice_id": id}

	err1 := collection.FindOne(context.TODO(), filter).Decode(&pay2)
	if err1 != nil{
		fmt.Println(err1)
		// we decode our body request params
		_ = json.NewDecoder(r.Body).Decode(&pay)

		// insert our book model.
		result, err := collection.InsertOne(context.TODO(), pay)

		if err != nil {
			helper.GetError(err, w)
			return
		}

		json.NewEncoder(w).Encode(result)
	} else {
		json.NewEncoder(w).Encode("Invoice can not be over paid")
	}

	fmt.Println(pay2.InvoiceId)
}

package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"fmt"
	"strconv"
	
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)
import "io/ioutil"
import "regexp"
import "invoice-crud/helper"
import "invoice-crud/models"
import "strings"

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

  	// set our port address
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

		// add item our array
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

	
	fmt.Println(json.NewDecoder(r.Body).Decode(&invoice))
	

	_ = json.NewDecoder(r.Body).Decode(&invoice)

	rate := getCurrency(invoice.Id)
	lnArr := []models.Line{}
	if s, err := strconv.ParseFloat(rate, 64); err == nil {
		fmt.Println(s) // 3.14159265
		for i := 0; i < len(invoice.Line); i++  {
			line := invoice.Line[i]
			if line.Currency != "CRC" {
				line.Price = s * line.Price
				line.Currency = "CRC"
			}
			lnArr = append(lnArr, line)
		}
	}
	invoice.Line = lnArr
	// connect db
	collection := helper.ConnectDB().Collection("invoices")

	// insert our book model.
	result, err := collection.InsertOne(context.TODO(), invoice)

	fmt.Println(invoice.Line)
	if err != nil {
		helper.GetError(err, w)
		return
	}

	json.NewEncoder(w).Encode(result)
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
	fmt.Println(json.NewDecoder(r.Body).Decode(&pay))

	_ = json.NewDecoder(r.Body).Decode(&pay)

	// connect db
	collection := helper.ConnectDB().Collection("payment")

	result, err := collection.InsertOne(context.TODO(), pay)

	if err != nil {
		helper.GetError(err, w)
		return
	}

	json.NewEncoder(w).Encode(result)
}


func getCurrency(invoice primitive.ObjectID) string {
	rtrn := ""
	resp, err := http.Get("https://gee.bccr.fi.cr/Indicadores/Suscripciones/WS/wsindicadoreseconomicos.asmx/ObtenerIndicadoresEconomicosXML?Indicador=318&FechaInicio=25/01/2019&FechaFinal=25/01/2019&Nombre=Maria&SubNiveles=N&CorreoElectronico=mariaobando09@gmail.com&Token=0OOIOO49MA")
	if err != nil {
		fmt.Println(err)
		return rtrn
	}
	fmt.Println(invoice)
	// fmt.Println(resp)
	defer resp.Body.Close()
	body, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		fmt.Println(err2)
		return rtrn
	}
	data := string(body)
	// fmt.Println(data)
	re := regexp.MustCompile(`&lt;NUM_VALOR&gt;(.*)&lt;/NUM_VALOR&gt;`)
	rplc1 := re.FindString(data)
	rplc2 := strings.Replace(rplc1, "&lt;NUM_VALOR&gt;", "", -1)
	rate := strings.Replace(rplc2, "&lt;/NUM_VALOR&gt;", "", -1)
	return rate
	// if s, err := strconv.ParseFloat(rate, 64); err == nil {
	// 	fmt.Println(s) // 3.14159265
	// 	return s
	// }
	// fmt.Println()
}
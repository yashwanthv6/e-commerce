package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Product struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Availability bool    `json:"availability"`
	Price        float64 `json:"price"`
	Category     string  `json:"category"`
}

type Order struct {
	ID           string  `json:"id"`
	ProductID    string  `json:"productId"`
	OrderValue   float64 `json:"orderValue"`
	DispatchDate string  `json:"dispatchDate,omitempty"`
	OrderStatus  string  `json:"orderStatus"`
	ProdQuantity int     `json:"prodQuantity"`
	IsPremium    bool    `json:"-"`
}

type Catalogue map[string]Product

type OrderList map[string]Order

var catalogue Catalogue
var orders OrderList

func getProductCatalogueHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(catalogue)
}

func placeOrderHandler(w http.ResponseWriter, r *http.Request) {
	var order Order
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if order.ProdQuantity <= 0 || order.ProdQuantity > 10 {
		http.Error(w, "Invalid product quantity", http.StatusBadRequest)
		return
	}

	product, ok := catalogue[order.ProductID]
	if !ok {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	order.OrderValue = float64(order.ProdQuantity) * product.Price

	if product.Category == "Premium" {
		order.OrderValue *= 0.9
		order.IsPremium = true
	}

	orders[order.ID] = order

	product.Availability = false
	catalogue[order.ProductID] = product

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func updateOrderStatusHandler(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("orderID")
	order, ok := orders[orderID]
	if !ok {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	var payload struct {
		Status       string `json:"status"`
		DispatchDate string `json:"dispatchDate,omitempty"`
	}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch payload.Status {
	case "Dispatched":
		if order.OrderStatus != "Placed" {
			http.Error(w, "Invalid order status transition", http.StatusBadRequest)
			return
		}
		order.OrderStatus = payload.Status
		order.DispatchDate = payload.DispatchDate
	case "Completed":
		if order.OrderStatus != "Dispatched" {
			http.Error(w, "Invalid order status transition", http.StatusBadRequest)
			return
		}
		order.OrderStatus = payload.Status
	case "Cancelled":
		if order.OrderStatus != "Placed" && order.OrderStatus != "Dispatched" {
			http.Error(w, "Invalid order status transition", http.StatusBadRequest)
			return
		}
		order.OrderStatus = payload.Status

	default:
		http.Error(w, "Invalid order status", http.StatusBadRequest)
		return
	}

	orders[orderID] = order
	json.NewEncoder(w).Encode(order)
}

func main() {
	catalogue = Catalogue{
		"1": {ID: "1", Name: "Product 1", Availability: true, Price: 10.0, Category: "Premium"},
		"2": {ID: "2", Name: "Product 2", Availability: true, Price: 5.0, Category: "Regular"},
		"3": {ID: "3", Name: "Product 3", Availability: true, Price: 3.0, Category: "Budget"},
	}

	orders = OrderList{}

	http.HandleFunc("/catalogue", getProductCatalogueHandler)
	http.HandleFunc("/order", placeOrderHandler)
	http.HandleFunc("/order/status", updateOrderStatusHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

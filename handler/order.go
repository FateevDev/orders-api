package handler

import (
	"fmt"
	"net/http"

	"github.com/FateevDev/orders-api/repository/order"
)

type Order struct{}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	order.RedisRepository{}
}

func (o *Order) List(w http.ResponseWriter, r *http.Request) {
	fmt.Println("List orders")
}

func (o *Order) Get(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get an order: ", r.URL.Path)
}

func (o *Order) Update(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Update an order")
}

func (o *Order) Delete(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete an order")
}

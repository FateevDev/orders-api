package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"

	"github.com/FateevDev/orders-api/model"
	orderRepo "github.com/FateevDev/orders-api/repository/order"
	"github.com/google/uuid"
)

type Order struct {
	Repository *orderRepo.RedisRepository
}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CustomerID uuid.UUID        `json:"customer_id"`
		LineItems  []model.LineItem `json:"line_items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()

	order := model.Order{
		OrderID:    rand.Uint64(),
		CustomerID: body.CustomerID,
		LineItems:  body.LineItems,
		CreatedAt:  &now,
	}

	err := o.Repository.Insert(r.Context(), order)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	marshal, err := json.Marshal(order)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(marshal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (o *Order) List(w http.ResponseWriter, r *http.Request) {
	cursorStr := r.URL.Query().Get("cursor")

	if cursorStr == "" {
		cursorStr = "0"
	}

	cursor, err := strconv.ParseUint(cursorStr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	const pageSize = 50

	all, err := o.Repository.FindAll(r.Context(), orderRepo.FindAllPage{
		Size:   pageSize,
		Offest: cursor,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type response struct {
		Items []model.Order `json:"items"`
		Next  uint64        `json:"next,omitempty"`
	}

	marshal, err := json.Marshal(response{
		Items: all.Orders,
		Next:  all.Cursor,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(marshal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (o *Order) Get(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get an order: ", r.URL.Path)
}

func (o *Order) Update(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Update an order")
}

func (o *Order) Delete(w http.ResponseWriter, r *http.Request) {
	err := o.Repository.Delete(r.Context(), 1)

	if errors.Is(err, orderRepo.ErrOrderWithIdNotFound(1)) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

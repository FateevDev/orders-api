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
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

type Order struct {
	Repository *orderRepo.RedisRepository
}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CustomerID uuid.UUID        `json:"customer_id" validate:"required"`
		LineItems  []model.LineItem `json:"line_items" validate:"required,min=1"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := validate.Struct(body)
	if err != nil {
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

	err = o.Repository.Insert(r.Context(), order)

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
	offset, err := getUint64QueryParameter(w, r, "offset", 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	limit, err := getUint64QueryParameter(w, r, "limit", 10)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if limit == 0 {
		http.Error(w, "limit must be greater than 0", http.StatusBadRequest)
		return
	}

	const maxLimit = 100

	if limit > maxLimit { // максимальный лимит
		limit = maxLimit
	}

	all, err := o.Repository.FindAll(r.Context(), orderRepo.FindAllPage{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type response struct {
		Items                []model.Order `json:"items"`
		orderRepo.Pagination `json:"meta"`
	}

	marshal, err := json.Marshal(response{
		Items:      all.Orders,
		Pagination: all.Pagination,
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
	idParam := chi.URLParam(r, "id")

	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order, err := o.Repository.FindById(r.Context(), id)
	if err != nil {
		if errors.As(err, orderRepo.ErrOrderWithIdNotFound(id)) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	marshal, err := json.Marshal(order)
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

func (o *Order) Update(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Update an order")
}

func (o *Order) Delete(w http.ResponseWriter, r *http.Request) {
	err := o.Repository.Delete(r.Context(), 1)

	if err == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if errors.Is(err, orderRepo.ErrOrderWithIdNotFound(1)) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func getUint64QueryParameter(w http.ResponseWriter, r *http.Request, queryParamName string, defaultValue uint64) (uint64, error) {
	valueStr := r.URL.Query().Get(queryParamName)

	if valueStr == "" {
		return defaultValue, nil
	}

	value, err := strconv.ParseUint(valueStr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return 0, nil
	}

	return value, nil
}

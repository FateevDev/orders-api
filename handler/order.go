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
	"github.com/FateevDev/orders-api/validation"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Order struct {
	Repository *orderRepo.RedisRepository
}

type FindAllPage struct {
	Offset uint64 `validate:"min=0"`
	Limit  uint64 `validate:"min=1,max=100"`
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

	err := validation.Validate.Struct(body)
	if err != nil {
		errorMessages := formatValidationErrors(err)
		response := map[string]interface{}{"errors": errorMessages}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
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
	params, hasValidationErrors := getPaginationParams(r, w)
	if hasValidationErrors {
		return
	}

	all, err := o.Repository.FindAll(r.Context(), orderRepo.FindAllPage{
		Limit:  params.Limit,
		Offset: params.Offset,
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
	id, err := getIdParam(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order, err := o.Repository.FindById(r.Context(), id)
	if err != nil {
		if errors.Is(err, orderRepo.ErrOrderNotFound) {
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
	id, err := getIdParam(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order, err := o.Repository.FindById(r.Context(), id)
	if err != nil {
		if errors.Is(err, orderRepo.ErrOrderNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var body struct {
		CustomerID uuid.UUID        `json:"customer_id" validate:"required"`
		LineItems  []model.LineItem `json:"line_items" validate:"required,min=1"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = validation.Validate.Struct(body)
	if err != nil {
		errorMessages := formatValidationErrors(err)
		response := map[string]interface{}{"errors": errorMessages}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	order.LineItems = body.LineItems

	err = o.Repository.Update(r.Context(), id, order)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	marshal, err := json.Marshal(order)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(marshal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (o *Order) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getIdParam(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = o.Repository.Delete(r.Context(), id)

	if err == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if errors.Is(err, orderRepo.ErrOrderNotFound) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func getIdParam(r *http.Request) (uint64, error) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid id parameter: %w", err)
	}

	return id, nil
}

func getPaginationParams(r *http.Request, w http.ResponseWriter) (FindAllPage, bool) {
	var params FindAllPage

	offsetStr := r.URL.Query().Get("offset")
	if offsetStr == "" {
		offsetStr = "0" // Default value
	}
	offset, _ := strconv.ParseUint(offsetStr, 10, 64)
	params.Offset = offset

	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limitStr = "10" // Default value
	}
	limit, _ := strconv.ParseUint(limitStr, 10, 64)
	params.Limit = limit

	err := validation.Validate.Struct(params)
	if err != nil {
		errorMessages := formatValidationErrors(err)
		response := map[string]interface{}{"errors": errorMessages}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return FindAllPage{}, true
	}
	return params, false
}

func formatValidationErrors(err error) map[string]string {
	errorsMap := make(map[string]string)

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		for _, e := range validationErrors {
			// The translator does all the work!
			errorsMap[e.Field()] = e.Translate(validation.Trans)
		}
	}
	return errorsMap
}

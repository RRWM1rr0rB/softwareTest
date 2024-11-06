package order

import (
	"encoding/json"
	"log"
	"net/http"

	policyOrder "software_test/internal/policy/order"
)

func (c *Controller) CreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input policyOrder.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	packs, err := c.orderPolicy.CreateOrder(ctx, input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Order created successfully: %+v", packs)

	response := policyOrder.CreateOrderResponse{Packs: packs.Packs}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

package api

import (
	"net/http"
)

func NewRouter(api API) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /products", api.Products)
	mux.HandleFunc("GET /products/{id}", api.Product)
	mux.HandleFunc("POST /availability", api.Availability)
	mux.HandleFunc("POST /bookings", api.CreateBooking)
	mux.HandleFunc("GET /bookings/{id}", api.Booking)
	mux.HandleFunc("POST /bookings/{id}/confirm", api.ConfirmBooking)

	return mux
}

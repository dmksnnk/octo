package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/dmksnnk/octo/internal"
	"github.com/dmksnnk/octo/internal/api"
	"github.com/dmksnnk/octo/internal/api/mocks"
	"github.com/dmksnnk/octo/internal/auth"
	"github.com/dmksnnk/octo/internal/platform"
	"github.com/dmksnnk/octo/internal/platform/golden"
	"github.com/stretchr/testify/mock"
)

var user = internal.User{
	ID:    456,
	Email: "test@example.com",
}

func TestAPIProducts(t *testing.T) {
	t.Run("products", func(t *testing.T) {
		products := []internal.Product{
			internal.ProductBase{
				ID:       "1",
				Name:     "Product 1",
				Capacity: 10,
			},
		}
		svc := mocks.NewMockService(t)
		svc.On("Products", mock.Anything, internal.CapabilityRequestNone).Return(products, nil)

		srv := newTestServer(t, svc)
		client := srv.Client()
		resp, err := client.Get(srv.URL + "/products")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assertEqualResponse(t, resp, http.StatusOK, golden.ReadBytes(t, "products.json"))
	})

	t.Run("products with price", func(t *testing.T) {
		products := []internal.Product{
			internal.ProductWithPrice{
				ProductBase: internal.ProductBase{
					ID:       "1",
					Name:     "Product 1",
					Capacity: 10,
				},
				CapabilityPrice: internal.CapabilityPrice{
					Price:    100,
					Currency: "EUR",
				},
			},
		}
		svc := mocks.NewMockService(t)
		svc.On("Products", mock.Anything, internal.CapabilityRequestPrice).Return(products, nil)
		srv := newTestServer(t, svc)
		client := srv.Client()

		req, err := http.NewRequest(http.MethodGet, srv.URL+"/products", http.NoBody)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Capability", "price")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assertEqualResponse(t, resp, http.StatusOK, golden.ReadBytes(t, "products-with-price.json"))
	})
}

func TestAPIProduct(t *testing.T) {
	t.Run("product", func(t *testing.T) {
		product := internal.ProductBase{
			ID:       "1",
			Name:     "Product 1",
			Capacity: 10,
		}
		svc := mocks.NewMockService(t)
		svc.On("Product", mock.Anything, 1, internal.CapabilityRequestNone).Return(product, nil)

		srv := newTestServer(t, svc)
		client := srv.Client()
		resp, err := client.Get(srv.URL + "/products/1")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assertEqualResponse(t, resp, http.StatusOK, golden.ReadBytes(t, "product.json"))
	})

	t.Run("invalid ID", func(t *testing.T) {
		svc := mocks.NewMockService(t)
		srv := newTestServer(t, svc)
		client := srv.Client()
		resp, err := client.Get(srv.URL + "/products/abc")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assertEqualResponse(t, resp, http.StatusBadRequest, golden.ReadBytes(t, "product-invalid-id.json"))
	})

	t.Run("product not found", func(t *testing.T) {
		svc := mocks.NewMockService(t)
		svc.On("Product", mock.Anything, 1, internal.CapabilityRequestNone).Return(nil, internal.ErrNotFound)

		srv := newTestServer(t, svc)
		client := srv.Client()
		resp, err := client.Get(srv.URL + "/products/1")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assertEqualResponse(t, resp, http.StatusNotFound, golden.ReadBytes(t, "product-not-found.json"))
	})
}

func TestAPIAvailability(t *testing.T) {
	t.Run("single availability", func(t *testing.T) {
		localDate := platform.Must(time.Parse("2006-01-02", "2025-01-20"))
		availability := internal.AvailabilityBase{
			ID:        "123",
			LocalDate: internal.Date(localDate),
			Status:    internal.AvailabilityStatusAvailable,
			Vacancies: 10,
			Available: true,
		}
		svc := mocks.NewMockService(t)
		svc.On("Availability", mock.Anything, 1, localDate, internal.CapabilityRequestNone).Return(availability, nil)
		srv := newTestServer(t, svc)
		client := srv.Client()
		resp, err := client.Post(srv.URL+"/availability", "application/json", golden.Open(t, "availability-request-single.json"))
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assertEqualResponse(t, resp, http.StatusOK, golden.ReadBytes(t, "availability-single.json"))
	})

	t.Run("single availability not found", func(t *testing.T) {
		localDate := platform.Must(time.Parse("2006-01-02", "2025-01-20"))
		svc := mocks.NewMockService(t)
		svc.On("Availability", mock.Anything, 1, localDate, internal.CapabilityRequestNone).Return(nil, internal.ErrNotFound)

		srv := newTestServer(t, svc)
		client := srv.Client()
		resp, err := client.Post(srv.URL+"/availability", "application/json", golden.Open(t, "availability-request-single.json"))
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assertEqualResponse(t, resp, http.StatusOK, []byte("[]\n"))
	})

	t.Run("single availability with price range", func(t *testing.T) {
		localDateStart := platform.Must(time.Parse("2006-01-02", "2025-01-20"))
		localDateEnd := platform.Must(time.Parse("2006-01-02", "2025-01-25"))
		availabilities := []internal.Availability{
			internal.AvailabilityWithPrice{
				AvailabilityBase: internal.AvailabilityBase{
					ID:        "123",
					LocalDate: internal.Date(localDateStart),
					Status:    internal.AvailabilityStatusAvailable,
					Vacancies: 10,
					Available: true,
				},
				CapabilityPrice: internal.CapabilityPrice{
					Price:    100,
					Currency: "EUR",
				},
			},
		}
		svc := mocks.NewMockService(t)
		svc.On("Availabilities", mock.Anything, 1, localDateStart, localDateEnd, internal.CapabilityRequestPrice).Return(availabilities, nil)

		srv := newTestServer(t, svc)
		client := srv.Client()
		req, err := http.NewRequest(http.MethodPost, srv.URL+"/availability", golden.Open(t, "availability-request-range.json"))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Capability", "price")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assertEqualResponse(t, resp, http.StatusOK, golden.ReadBytes(t, "availability-range-with-price.json"))
	})
}

func TestAPIBooking(t *testing.T) {
	booking := internal.BookingBase{
		ID:             "123",
		ProductID:      "1",
		AvailabilityID: "123",
		Status:         internal.BookingStatusReserved,
		Units: []internal.Unit{
			internal.UnitBase{
				ID:     "1",
				Ticket: platform.ToPtr("ticket 1"),
			},
		},
	}
	bookingWithPrice := internal.BookingWithPrice{
		BookingBase: internal.BookingBase{
			ID:             "123",
			ProductID:      "1",
			AvailabilityID: "123",
			Status:         internal.BookingStatusReserved,
			Units: []internal.Unit{
				internal.UnitWithPrice{
					UnitBase: internal.UnitBase{
						ID:     "1",
						Ticket: platform.ToPtr("ticket 1"),
					},
					CapabilityPrice: internal.CapabilityPrice{
						Price:    100,
						Currency: "EUR",
					},
				},
			},
		},
		CapabilityPrice: internal.CapabilityPrice{
			Price:    100,
			Currency: "EUR",
		},
	}

	t.Run("get booking", func(t *testing.T) {
		svc := mocks.NewMockService(t)
		svc.On("Booking", mock.Anything, 123, user.ID, internal.CapabilityRequestNone).Return(booking, nil)
		srv := newTestServer(t, svc)

		client := srv.Client()
		resp, err := client.Get(srv.URL + "/bookings/123")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assertEqualResponse(t, resp, http.StatusOK, golden.ReadBytes(t, "booking.json"))
	})

	t.Run("get booking with price", func(t *testing.T) {
		svc := mocks.NewMockService(t)
		svc.On("Booking", mock.Anything, 123, user.ID, internal.CapabilityRequestPrice).Return(bookingWithPrice, nil)
		srv := newTestServer(t, svc)

		client := srv.Client()
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/bookings/123", http.NoBody)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Capability", "price")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assertEqualResponse(t, resp, http.StatusOK, golden.ReadBytes(t, "booking-with-price.json"))
	})

	t.Run("create booking", func(t *testing.T) {
		svc := mocks.NewMockService(t)
		params := internal.CreateBookingRequest{
			ProductID:      1,
			AvailabilityID: 123,
			Units:          1,
			UserID:         user.ID,
		}
		svc.On("CreateBooking", mock.Anything, params).Return(123, nil)
		svc.On("Booking", mock.Anything, 123, user.ID, internal.CapabilityRequestNone).Return(booking, nil)
		srv := newTestServer(t, svc)

		client := srv.Client()
		resp, err := client.Post(srv.URL+"/bookings", "application/json", golden.Open(t, "booking-create-request.json"))
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assertEqualResponse(t, resp, http.StatusOK, golden.ReadBytes(t, "booking.json"))
	})

	t.Run("create booking conflict", func(t *testing.T) {
		svc := mocks.NewMockService(t)
		params := internal.CreateBookingRequest{
			ProductID:      1,
			AvailabilityID: 123,
			Units:          1,
			UserID:         user.ID,
		}
		svc.On("CreateBooking", mock.Anything, params).Return(0, internal.ErrNotAvailable)
		srv := newTestServer(t, svc)

		client := srv.Client()
		resp, err := client.Post(srv.URL+"/bookings", "application/json", golden.Open(t, "booking-create-request.json"))
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assertEqualResponse(t, resp, http.StatusConflict, golden.ReadBytes(t, "booking-conflict.json"))
	})

	t.Run("confirm booking", func(t *testing.T) {
		svc := mocks.NewMockService(t)
		svc.On("ConfirmBooking", mock.Anything, 123, user.ID).Return(nil)
		svc.On("Booking", mock.Anything, 123, user.ID, internal.CapabilityRequestNone).Return(booking, nil)
		srv := newTestServer(t, svc)

		client := srv.Client()
		resp, err := client.Post(srv.URL+"/bookings/123/confirm", "application/json", golden.Open(t, "booking-create-request.json"))
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assertEqualResponse(t, resp, http.StatusOK, golden.ReadBytes(t, "booking.json"))
	})
}

func newTestServer(t *testing.T, svc api.Service) *httptest.Server {
	t.Helper()

	router := api.NewRouter(api.NewAPI(svc))
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := auth.ContextWithUser(r.Context(), user)
			router.ServeHTTP(w, r.WithContext(ctx))
		}),
	)
	t.Cleanup(srv.Close)

	return srv
}

func assertEqualResponse(t *testing.T, resp *http.Response, wantStatus int, wantBody []byte) {
	t.Helper()

	if resp.StatusCode != wantStatus {
		t.Errorf("want status %d, got %d", wantStatus, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	assertEqualJSON(t, body, wantBody)
}

func assertEqualJSON(t *testing.T, respBody, wantBody []byte) {
	t.Helper()

	var want, got any
	if err := json.Unmarshal(respBody, &got); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	if err := json.Unmarshal(wantBody, &want); err != nil {
		t.Fatalf("failed to unmarshal expected body: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want: %v, got: %v", want, got)
	}
}

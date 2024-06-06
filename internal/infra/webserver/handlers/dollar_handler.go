package handlers

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/matheusmhmelo/FullCycle-client-server-api/internal/entity"
	"github.com/matheusmhmelo/FullCycle-client-server-api/internal/infra/database"
	"io"
	"log"
	"net/http"
	"time"
)

type DollarHandler struct {
	DollarDB database.DollarInterface
}

func NewDollarHandler(db database.DollarInterface) *DollarHandler {
	return &DollarHandler{
		DollarDB: db,
	}
}

func (h *DollarHandler) GetDollar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	reqChan := make(chan *http.Request)
	errChan := make(chan error)
	go func() {
		req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
		if err != nil {
			errChan <- err
		}
		reqChan <- req
	}()

	var req *http.Request
	select {
	case <-time.After(200 * time.Millisecond):
		log.Println("Request timeout")
		w.WriteHeader(http.StatusRequestTimeout)
		return
	case <-errChan:
		w.WriteHeader(http.StatusBadRequest)
		return
	case req = <-reqChan:
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var result entity.DollarResponse
	if err = json.Unmarshal(body, &result); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	go func() {
		err = h.DollarDB.Create(ctx, &entity.Dollar{
			ID:        uuid.New().String(),
			Value:     result.Currency.Value,
			CreatedAt: time.Now(),
		})
		errChan <- err
	}()

	select {
	case <-time.After(10 * time.Millisecond):
		log.Println("Database timeout")
		w.WriteHeader(http.StatusRequestTimeout)
		return
	case err = <-errChan:
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result.Currency.Value))
}

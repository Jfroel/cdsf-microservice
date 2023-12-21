package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Jfroel/cdsf-microservice/proto/filter"
)

type Proxy struct {
	port         int
	filterClient filter.FilterServiceClient
	ID           string
}

// NewFrontend creates a new Frontend instance with the specified configuration.
func NewProxy(port int, filterAddr, ID string) *Proxy {
	p := &Proxy{
		port:         port,
		filterClient: filter.NewFilterServiceClient(dial(filterAddr)),
		ID:           ID,
	}
	return p
}

func (s *Proxy) Run() error {
	// http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/insert", s.insertHandler)
	http.HandleFunc("/get-max", s.getMaxHandler)
	http.HandleFunc("/get-min", s.getMinHandler)
	http.HandleFunc("/remove-max", s.removeMaxHandler)
	http.HandleFunc("/remove-min", s.removeMinHandler)
	http.HandleFunc("/get-size", s.getSizeHandler)
	http.HandleFunc("/clear", s.clearHandler)

	log.Printf("http to grpc proxy %v server running at port: %d", s.ID, s.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
}

func (s *Proxy) insertHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	ctx := r.Context()

	scoreStr := r.URL.Query().Get("score")

	if scoreStr == "" {
		http.Error(w, "Malformed request to `/insert` endpoint!", http.StatusBadRequest)
		return
	}
	score, err := strconv.ParseFloat(scoreStr, 32)
	if err != nil {
		http.Error(w, "Malformed request to `/insert` endpoint!", http.StatusBadRequest)
	}

	req := &filter.InsertItemRequest{
		Item: &filter.FilterItem{
			Score: float32(score),
			Data:  []byte{0x01, 0x02, 0x03, 0x04},
		},
	}
	reply, err := s.filterClient.InsertItem(ctx, req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate the duration in microseconds
	duration := int64(time.Since(start).Microseconds())
	in, _ := json.Marshal(req)
	// out, _ := json.Marshal(reply)
	inStr, outStr := string(in), "{}"

	errStr := fmt.Sprintf("%v", err)
	if err == nil {
		errStr = "<nil>"
	}

	logMsg("proxy.insertHandler", inStr, outStr, errStr, duration)

	err = json.NewEncoder(w).Encode(reply)
}

func (s *Proxy) getMaxHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	ctx := r.Context()

	req := &filter.GetMaxItemRequest{}
	reply, err := s.filterClient.GetMaxItem(ctx, req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate the duration in microseconds
	duration := int64(time.Since(start).Microseconds())
	in, _ := json.Marshal(req)
	// out, _ := json.Marshal(reply)
	inStr, outStr := string(in), "{}"

	errStr := fmt.Sprintf("%v", err)
	if err == nil {
		errStr = "<nil>"
	}

	logMsg("proxy.getMaxHandler", inStr, outStr, errStr, duration)

	err = json.NewEncoder(w).Encode(reply)
}

func (s *Proxy) getMinHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	ctx := r.Context()

	req := &filter.GetMinItemRequest{}
	reply, err := s.filterClient.GetMinItem(ctx, req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate the duration in microseconds
	duration := int64(time.Since(start).Microseconds())
	in, _ := json.Marshal(req)
	// out, _ := json.Marshal(reply)
	inStr, outStr := string(in), "{}"

	errStr := fmt.Sprintf("%v", err)
	if err == nil {
		errStr = "<nil>"
	}

	logMsg("proxy.getMinHandler", inStr, outStr, errStr, duration)

	err = json.NewEncoder(w).Encode(reply)
}

func (s *Proxy) removeMaxHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	ctx := r.Context()

	req := &filter.RemoveMaxItemRequest{}
	reply, err := s.filterClient.RemoveMaxItem(ctx, req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate the duration in microseconds
	duration := int64(time.Since(start).Microseconds())
	in, _ := json.Marshal(req)
	// out, _ := json.Marshal(reply)
	inStr, outStr := string(in), "{}"

	errStr := fmt.Sprintf("%v", err)
	if err == nil {
		errStr = "<nil>"
	}

	logMsg("proxy.removeMaxHandler", inStr, outStr, errStr, duration)

	err = json.NewEncoder(w).Encode(reply)
}

func (s *Proxy) removeMinHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	ctx := r.Context()

	req := &filter.RemoveMinItemRequest{}
	reply, err := s.filterClient.RemoveMinItem(ctx, req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate the duration in microseconds
	duration := int64(time.Since(start).Microseconds())
	in, _ := json.Marshal(req)
	// out, _ := json.Marshal(reply)
	inStr, outStr := string(in), "{}"

	errStr := fmt.Sprintf("%v", err)
	if err == nil {
		errStr = "<nil>"
	}

	logMsg("proxy.removeMinHandler", inStr, outStr, errStr, duration)

	err = json.NewEncoder(w).Encode(reply)
}

func (s *Proxy) getSizeHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	ctx := r.Context()

	req := &filter.GetSizeRequest{}
	reply, err := s.filterClient.GetSize(ctx, req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate the duration in microseconds
	duration := int64(time.Since(start).Microseconds())
	in, _ := json.Marshal(req)
	// out, _ := json.Marshal(reply)
	inStr, outStr := string(in), "{}"

	errStr := fmt.Sprintf("%v", err)
	if err == nil {
		errStr = "<nil>"
	}

	logMsg("proxy.getSizeHandler", inStr, outStr, errStr, duration)

	err = json.NewEncoder(w).Encode(reply)
}

func (s *Proxy) clearHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	ctx := r.Context()

	req := &filter.ClearRequest{}
	reply, err := s.filterClient.Clear(ctx, req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate the duration in microseconds
	duration := int64(time.Since(start).Microseconds())
	in, _ := json.Marshal(req)
	// out, _ := json.Marshal(reply)
	inStr, outStr := string(in), "{}"

	errStr := fmt.Sprintf("%v", err)
	if err == nil {
		errStr = "<nil>"
	}

	logMsg("proxy.clearHandler", inStr, outStr, errStr, duration)

	err = json.NewEncoder(w).Encode(reply)
}

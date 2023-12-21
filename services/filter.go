package services

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/Jfroel/cdsf-microservice/apps"
	"github.com/Jfroel/cdsf-microservice/proto/filter"
	"google.golang.org/grpc"
)

type Filter struct {
	name string
	port int
	filter.FilterServiceServer
	app apps.ConcurrentDataStreamFilter
}

func NewFilter(name string, port int, lkType string, capacity int) *Filter {
	return &Filter{
		name: name,
		port: port,
		app:  apps.NewCDSFApp(lkType, capacity),
	}
}

func (s *Filter) Run() error {
	// Create a new gRPC server instance.
	srv := grpc.NewServer()

	// Register the Cache server implementation with the gRPC server.
	filter.RegisterFilterServiceServer(srv, s)

	// Create a TCP listener that listens for incoming requests on the specified port.
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// (Optional) Log a message indicating that the server is running and listening on the specified port.
	log.Printf("filter server <%s> running at port: %d", s.name, s.port)
	return srv.Serve(lis)
}

func (s *Filter) InsertItem(ctx context.Context, req *filter.InsertItemRequest) (*filter.InsertItemResponse, error) {
	resp := &filter.InsertItemResponse{Success: true}
	item := req.GetItem()
	err := s.app.Insert(item)
	if err != nil {
		resp.Success = false
	}
	return resp, err
}

func (s *Filter) GetMaxItem(ctx context.Context, req *filter.GetMaxItemRequest) (*filter.GetMaxItemResponse, error) {
	resp := &filter.GetMaxItemResponse{}
	item, err := s.app.GetMax()
	if err != nil {
		return resp, err
	}
	resp.Item = item
	return resp, err
}

func (s *Filter) GetMinItem(ctx context.Context, req *filter.GetMinItemRequest) (*filter.GetMinItemResponse, error) {
	resp := &filter.GetMinItemResponse{}
	item, err := s.app.GetMin()
	if err != nil {
		return resp, err
	}
	resp.Item = item
	return resp, err
}

func (s *Filter) RemoveMaxItem(ctx context.Context, req *filter.RemoveMaxItemRequest) (*filter.RemoveMaxItemResponse, error) {
	resp := &filter.RemoveMaxItemResponse{}
	item, err := s.app.RemoveMax()
	if err != nil {
		return resp, err
	}
	resp.Item = item
	return resp, err
}

func (s *Filter) RemoveMinItem(ctx context.Context, req *filter.RemoveMinItemRequest) (*filter.RemoveMinItemResponse, error) {
	resp := &filter.RemoveMinItemResponse{}
	item, err := s.app.RemoveMin()
	if err != nil {
		return resp, err
	}
	resp.Item = item
	return resp, err
}

func (s *Filter) GetSize(ctx context.Context, req *filter.GetSizeRequest) (*filter.GetSizeResponse, error) {
	resp := &filter.GetSizeResponse{}
	size := s.app.GetSize()

	resp.Size = int32(size)
	return resp, nil
}

func (s *Filter) Clear(ctx context.Context, req *filter.ClearRequest) (*filter.ClearResponse, error) {
	resp := &filter.ClearResponse{Success: true}
	err := s.app.Clear()
	if err != nil {
		resp.Success = false
	}
	return resp, err
}

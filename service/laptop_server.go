package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"laptop-app-using-grpc/pb/pb"
	"log"
)

//Laptop server which provides the services
type LaptopServer struct {
	Store LaptopStore
}

//Returning a new laptop server
func NewLaptopServer(store LaptopStore) *LaptopServer {
	return &LaptopServer{store}
}

//Creating the unary RPC to create a new laptop
func (server *LaptopServer) CreateLaptop(ctx context.Context, req *pb.CreateLaptopRequest) (*pb.CreateLaptopResponse, error) {
	laptop := req.GetLaptop()
	log.Println("received a create laptop request using id: ", laptop.Id)

	if len(laptop.Id) > 0 {
		//Checking for valid UUID
		_, err := uuid.Parse(laptop.Id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Laptop ID is not valid uuid: %v", err)
		}
	} else {

		id, err := uuid.NewRandom()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Cannot generate a new laptop Id: %v", err)
		}

		laptop.Id = id.String()
	}

	//Save the laptop to in-memory store
	err := server.Store.Save(laptop)
	if err != nil {
		code := codes.Internal
		if errors.Is(err, ErrorAlreadyExists) {
			code = codes.AlreadyExists
		}

		return nil, status.Errorf(code, "Cannot save laptop to store: %v", err)
	}

	log.Println("Saved laptop with id: ", laptop.Id)

	res := &pb.CreateLaptopResponse{
		Id: laptop.Id,
	}
	return res, nil
}

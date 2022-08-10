package service

import (
	"bytes"
	"context"
	"errors"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"laptop-app-using-grpc/pb/pb"
	"log"
)

// maximum 1 megabyte
const maxImageSize = 1 << 20

// LaptopServer which provides the services
type LaptopServer struct {
	laptopStore LaptopStore
	imageStore  ImageStore
}

// NewLaptopServer Returning a new laptop server
func NewLaptopServer(laptopStore LaptopStore, imageStore ImageStore) *LaptopServer {
	return &LaptopServer{laptopStore, imageStore}
}

// CreateLaptop Creating the unary RPC to create a new laptop
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

	//Time for heavy processing
	/*
		time.Sleep(6 * time.Second)
	*/

	//For cancelled and time out scenarios
	if ctx.Err() == context.Canceled {
		log.Print("Request is canceled")
		return nil, status.Error(codes.Canceled, "Request is canceled")
	}

	if ctx.Err() == context.DeadlineExceeded {
		log.Println("Deadline is extended")
		return nil, status.Error(codes.DeadlineExceeded, "Deadline is exceeded")
	}

	//Save the laptop to in-memory store
	err := server.laptopStore.Save(laptop)
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

// SearchLaptop is server-streaming RPC to seach for laptops
func (server *LaptopServer) SearchLaptop(req *pb.SearchLaptopRequest, stream pb.LaptopService_SearchLaptopServer) (outErr error) {
	filter := req.GetFilter()
	log.Printf("Recieve a search laptop request with filter: %v", filter)

	err := server.laptopStore.Search(
		stream.Context(),
		filter,
		func(laptop *pb.Laptop) {
			res := &pb.SearchLaptopResponse{Laptop: laptop}

			err := stream.Send(res)
			if err != nil {
				outErr = status.Errorf(codes.Unknown, "cannot send response: %v", err)
				return
			}

			log.Printf("Send laptop with id: %s", laptop.GetId())
		})
	if err != nil {
		return status.Errorf(codes.Internal, "unexpected error: %v", err)
	}

	return nil
}

//UploadImage is client-streaming RPC to upload laptop Images
func (server *LaptopServer) UploadImage(stream pb.LaptopService_UploadImageServer) error {
	req, err := stream.Recv()
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "Cannot receive image info"))
	}

	laptopID := req.GetInfo().GetLaptopId()
	imageType := req.GetInfo().GetImageType()
	log.Printf("Received an upload image request for laptop %s with image type %s", laptopID, imageType)

	laptop, err := server.laptopStore.Find(laptopID)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "Cannot find laptop: %v", err))
	}

	if laptop != nil {
		return logError(status.Errorf(codes.Internal, "Laptop %s does not exist", laptopID))
	}

	imageData := bytes.Buffer{}
	imageSize := 0

	for {
		log.Printf("Waiting to recieve more data")

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("No more data")
			break
		}
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "Cannot receive chunk data: %v", err))
		}

		chunk := req.GetChunkData()
		size := len(chunk)

		imageSize += size
		if imageSize > maxImageSize {
			return logError(status.Errorf(codes.InvalidArgument, "image is too large: %d > %d", imageSize, maxImageSize))
		}

		_, err = imageData.Write(chunk)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "Cannot write chunk data: %v", err))
		}
	}

	imageID, err := server.imageStore.Save(laptopID, imageType, imageData)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "Cannot save image to the store: %v", err))
	}

	res := &pb.UploadImageResponse{
		Id:   imageID,
		Size: uint32(imageSize),
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "Cannot send response: %v", err))
	}

	log.Printf("Saved image with id: %s, size: %d", imageID, imageSize)
	return nil
}

//Utility function for logging errors to console
func logError(err error) error {
	if err != nil {
		log.Print(err)
	}

	return err
}

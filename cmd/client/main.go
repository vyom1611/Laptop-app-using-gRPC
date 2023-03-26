package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"io"
	"laptop-app-using-grpc/pb/pb"
	"laptop-app-using-grpc/sample"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func createLaptop(laptopClient pb.LaptopServiceClient, laptop *pb.Laptop) {
	//Removing the generated universal Id with every laptop
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	//Set timeout for connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := laptopClient.CreateLaptop(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			log.Println("Laptop Already Exists")
		} else {
			//Worse
			log.Fatal("Cannot create Laptop", err)
		}
		return
	}
	log.Printf("Created Laptop with Id: %s", res.Id)
}

func rateLaptop(laptopClient pb.LaptopServiceClient, laptopIDs []string, scores []float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := laptopClient.RateLaptop(ctx)
	if err != nil {
		return fmt.Errorf("cannot rate laptop: %v", err)
	}

	waitResponse := make(chan error)
	// go routine to receive responses
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				log.Print("no more responses")
				waitResponse <- nil
				return
			}
			if err != nil {
				waitResponse <- fmt.Errorf("cannot receive stream response: %v", err)
				return
			}

			log.Print("received response: ", res)
		}
	}()

	// send requests
	for i, laptopID := range laptopIDs {
		req := &pb.RateLaptopRequest{
			LaptopId: laptopID,
			Score:    scores[i],
		}

		err := stream.Send(req)
		if err != nil {
			return fmt.Errorf("cannot send stream request: %v - %v", err, stream.RecvMsg(nil))
		}

		log.Print("sent request: ", req)
	}

	err = stream.CloseSend()
	if err != nil {
		return fmt.Errorf("cannot close send: %v", err)
	}

	err = <-waitResponse
	return err
}

func searchLaptop(laptopClient pb.LaptopServiceClient, filter *pb.Filter) {
	log.Print("search filter: ", filter)

	// Setting it up so that when the context deadline is met, all resources are released
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Making a search request with the parameters of the filter
	req := &pb.SearchLaptopRequest{Filter: filter}
	// Sending the search request to the client in context of the search laptop stream
	stream, err := laptopClient.SearchLaptop(ctx, req)
	if err != nil {
		log.Fatal("cannot search laptop: ", err)
	}

	for {
		res, err := stream.Recv()

		// Gracefully ending reading input files
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatal("cannot receive response: ", err)
		}

		// Printing the details of the response laptop configurations
		laptop := res.GetLaptop()
		log.Print("- found: ", laptop.GetId())
		log.Print(" + brand: ", laptop.GetBrand())
		log.Print(" + name: ", laptop.GetName())
		log.Print(" + cpu cores: ", laptop.GetCpu().GetCpuCores())
		log.Print(" + cpu min ghz: ", laptop.GetCpu().GetMinGhz())
		log.Print(" + ram: ", laptop.GetRam())
		log.Print(" + price: ", laptop.GetPriceUsd(), "usd")
	}
}

// uploadImage function on client-side
func uploadImage(laptopClient pb.LaptopServiceClient, laptopID string, imagePath string) {
	// Opens the imagePath in the arguments for reading
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatal("Cannot open image file: ", err)
	}

	// Setting up timeout so that when the context deadline is met, all resources are released
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Defining the uploadImage stream from client-side in the context
	stream, err := laptopClient.UploadImage(ctx)
	if err != nil {
		log.Fatal("Cannot upload laptop image file: ", err)
	}

	// Constructing the upload Image request model with the data of the request_info (like laptop id and image type)
	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId:  laptopID,
				ImageType: filepath.Ext(imagePath),
			},
		},
	}

	// Sending the request into the client-side stream
	err = stream.Send(req)
	if err != nil {
		log.Fatal("Cannot send image info: ", err)
	}

	// Reading the open file at the imagePath and making a buffer of bytes
	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("Cannot read chunk to buffer: ", err)
		}

		// Writing uploadImage request to the bytes buffer in the open file path
		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		// Sending the new request consisting og bytes chunk to stream
		err = stream.Send(req)
		if err != nil {
			log.Fatal("Cannot send chunk to server: ", err, stream.RecvMsg(nil))
		}
	}

	// Closing the client-side stream after receiving the request and sending out a response
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("Cannot receive response: ", err)
	}

	// The response prints out the laptop Id and image size
	log.Printf("Image uploaded with ID: %s, size: %d", res.GetId(), res.GetSize())
}

// testCreateLaptop tests the createLaptop method on client-side
func testCreateLaptop(laptopClient pb.LaptopServiceClient) {
	createLaptop(laptopClient, sample.NewLaptop())
}

// testUploadImage tests the uploadImage function by creating a new laptopClient and sending a sample laptop to uploadImage
func testUploadImage(laptopClient pb.LaptopServiceClient) {
	laptop := sample.NewLaptop()
	createLaptop(laptopClient, laptop)
	uploadImage(laptopClient, laptop.GetId(), "tmp/laptop.jpg")
}

// testSearchLaptop creates laptop with filter and calls searchLaptop with the defined filter
func testSearchLaptop(laptopClient pb.LaptopServiceClient) {
	for i := 0; i < 10; i++ {
		createLaptop(laptopClient, sample.NewLaptop())
	}

	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz:   2.5,
		MinRam:      &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE},
	}

	searchLaptop(laptopClient, filter)
}

func testRateLaptop(laptopClient pb.LaptopServiceClient) {
	n := 3
	laptopIDs := make([]string, n)

	for i := 0; i < n; i++ {
		laptop := sample.NewLaptop()
		laptopIDs[i] = laptop.GetId()
		createLaptop(laptopClient, laptop)
	}

	scores := make([]float64, n)
	for {
		fmt.Print("rate laptop (y/n)? ")
		var answer string
		fmt.Scan(&answer)

		if strings.ToLower(answer) != "y" {
			break
		}

		for i := 0; i < n; i++ {
			scores[i] = sample.RandomLaptopScore()
		}

		err := rateLaptop(laptopClient, laptopIDs, scores)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	serverAddress := flag.String("address", "", "the server address")
	flag.Parse()
	log.Printf("dial server %s", *serverAddress)

	conn, err := grpc.Dial(*serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("cannot dial server: ", err)
	}

	//Starting laptop service (client-side) on grpc connection
	laptopClient := pb.NewLaptopServiceClient(conn)

	//for i := 0; i < 10; i++ {
	//	createLaptop(laptopClient, sample.NewLaptop())
	//}
	//
	//filter := &pb.Filter{
	//	MaxPriceUsd: 3000,
	//	MinCpuCores: 4,
	//	MinCpuGhz:   2.5,
	//	MinRam:      &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE},
	//}
	//
	//searchLaptop(laptopClient, filter)
	//testUploadImage(laptopClient)
	testRateLaptop(laptopClient)
}

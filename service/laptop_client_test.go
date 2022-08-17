package service_test

import (
	"bufio"
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"laptop-app-using-grpc/pb/pb"
	"laptop-app-using-grpc/sample"
	"laptop-app-using-grpc/serializer"
	"laptop-app-using-grpc/service"
	"net"
	"os"
	"path/filepath"
	"testing"
)

// Unit-Tests for all RPCs created and used on client-side

func TestClientCreateLaptop(t *testing.T) {
	t.Parallel()

	laptopStore := service.NewInMemoryLaptopStore()
	serverAddress := startTestLaptopServer(t, laptopStore, nil)
	laptopClient := newTestLaptopClient(t, serverAddress)

	laptop := sample.NewLaptop()
	expectedID := laptop.Id
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	res, err := laptopClient.CreateLaptop(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, expectedID, res.Id)

	//Checking that laptop is installed on the server/store using response id
	other, err := laptopStore.Find(res.Id)
	require.NoError(t, err)
	require.NotNil(t, other)

	//Checking that the saved laptop equals the laptop in response
	requireSameLaptop(t, laptop, other)
}

func TestClientSearchLaptop(t *testing.T) {
	t.Parallel()

	// defining an example filter
	filter := &pb.Filter{
		MaxPriceUsd: 2000,
		MinCpuGhz:   2.2,
		MinCpuCores: 4,
		MinRam:      &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE},
	}

	// creating a new Laptop Store to test with
	laptopStore := service.NewInMemoryLaptopStore()
	expectedIDs := make(map[string]bool)

	// Creating 6 sample new laptops with different configs for each case
	for i := 0; i < 6; i++ {
		laptop := sample.NewLaptop()

		switch i {
		case 0:
			laptop.PriceUsd = 2500
		case 1:
			laptop.Cpu.CpuCores = 2
		case 2:
			laptop.Cpu.MinGhz = 2.0
		case 3:
			laptop.Ram = &pb.Memory{
				Value: 4096,
				Unit:  pb.Memory_MEGABYTE,
			}
		case 4:
			laptop.PriceUsd = 1999
			laptop.Cpu.CpuCores = 4
			laptop.Cpu.MinGhz = 2.5
			laptop.Cpu.MaxGhz = laptop.Cpu.MinGhz + 2.0
			laptop.Ram = &pb.Memory{
				Value: 16,
				Unit:  pb.Memory_GIGABYTE,
			}
			expectedIDs[laptop.Id] = true
		case 5:
			laptop.PriceUsd = 2000
			laptop.Cpu.CpuCores = 6
			laptop.Cpu.MinGhz = 2.8
			laptop.Cpu.MaxGhz = laptop.Cpu.MinGhz + 2.0
			laptop.Ram = &pb.Memory{
				Value: 64,
				Unit:  pb.Memory_GIGABYTE,
			}
			expectedIDs[laptop.Id] = true
		}

		// Saving the generated laptops to store
		err := laptopStore.Save(laptop)
		require.NoError(t, err)
	}

	serverAddress := startTestLaptopServer(t, laptopStore, nil)
	laptopClient := newTestLaptopClient(t, serverAddress)

	req := &pb.SearchLaptopRequest{Filter: filter}
	stream, err := laptopClient.SearchLaptop(context.Background(), req)
	require.NoError(t, err)

	//Laptops found
	found := 0
	for {
		// Receiving data in stream
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}

		require.NoError(t, err)
		// Testing if expectedIDs are in the generated laptops data
		require.Contains(t, expectedIDs, res.GetLaptop().GetId())

		// Increasing count for correct findings
		found += 1
	}

	// Checking if total found laptops are equal to length of expected Ids
	require.Equal(t, len(expectedIDs), found)
}

func TestClientUploadImage(t *testing.T) {
	t.Parallel()

	// All temporary data in tests go to /tmp folder
	testImageFolder := "../tmp"

	// Creating a laptop and image store for testing
	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore(testImageFolder)

	laptop := sample.NewLaptop()

	// Saving a new generated laptop to store
	err := laptopStore.Save(laptop)
	require.NoError(t, err)

	// Starting laptop server
	serverAddress := startTestLaptopServer(t, laptopStore, imageStore)
	laptopClient := newTestLaptopClient(t, serverAddress)

	// Defining image path in tmp folder
	imgPath := fmt.Sprintf("%s/laptop.jpg", testImageFolder)
	file, err := os.Open(imgPath)
	require.NoError(t, err)
	defer file.Close()

	// Using uploadImage RPC on a new stream of this context
	stream, err := laptopClient.UploadImage(context.Background())
	require.NoError(t, err)

	// Imagetype is the file extension
	imageType := filepath.Ext(imgPath)

	// making a request model
	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId:  laptop.GetId(),
				ImageType: imageType,
			},
		},
	}

	// Sending the request in the stream
	err = stream.Send(req)
	require.NoError(t, err)

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)
	size := 0

	for {
		// Checking for bytes buffer which are read
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		// Increasing size of file in correspondence to the buffer length
		size += n

		// Making the created chunks of buffer into new request
		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		// Sending the new request to the stream
		err = stream.Send(req)
		require.NoError(t, err)
	}

	// Closing the stream after receiving data and sending a response
	res, err := stream.CloseAndRecv()

	// Checking for errors from the response
	require.NoError(t, err)
	require.NotZero(t, laptop.GetId())
	require.EqualValues(t, size, res.GetSize())

	savedImagePath := fmt.Sprintf("%s/%s%s", testImageFolder, res.GetId(), imageType)

	require.FileExists(t, savedImagePath)
	require.NoError(t, os.Remove(savedImagePath))
}

func startTestLaptopServer(t *testing.T, laptopStore service.LaptopStore, imageStore service.ImageStore) string {
	laptopServer := service.NewLaptopServer(laptopStore, imageStore)

	//Creating a server using grpc
	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	//Establishing a tcp connection on a random available port
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	//non blocking call
	go grpcServer.Serve(listener)

	//Returning the laptop server with the address on the tco connection as a string
	return listener.Addr().String()
}

func newTestLaptopClient(t *testing.T, serverAddress string) pb.LaptopServiceClient {
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	return pb.NewLaptopServiceClient(conn)
}

func requireSameLaptop(t *testing.T, laptop1 *pb.Laptop, laptop2 *pb.Laptop) {
	/* Converting to json to compare them since the protobuf messages has some unique generated
	methods and would not work for the comparison */
	json1, err := serializer.ProtobufToJSON(laptop1)
	require.NoError(t, err)

	json2, err := serializer.ProtobufToJSON(laptop2)
	require.NoError(t, err)

	require.Equal(t, json1, json2)
}

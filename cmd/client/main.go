package main

import (
	"context"
	"flag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"laptop-app-using-grpc/pb/pb"
	"laptop-app-using-grpc/sample"
	"log"
	"time"
)

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

	laptop := sample.NewLaptop()

	//Removing the generated universal Id with every laptop
	laptop.Id = ""
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	//Set timeout for connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := laptopClient.CreateLaptop(context.Background(), req)
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

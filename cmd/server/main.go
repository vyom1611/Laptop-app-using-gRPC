package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"laptop-app-using-grpc/pb/pb"
	"laptop-app-using-grpc/service"
	"log"
	"net"
)

func main() {
	port := flag.Int("port", 0, "the server port")
	flag.Parse()
	log.Printf("The server started on port %d", *port)

	//Defining stores
	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("img")
	ratingStore := service.NewInMemoryRatingStore()
	//Creating a laptop server service
	laptopServer := service.NewLaptopServer(laptopStore, imageStore, ratingStore)
	//Creating a grpc web server
	grpcServer := grpc.NewServer()
	//Adding laptop server in grpc service
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	reflection.Register(grpcServer)

	address := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("Cannot start server: ", err)
	}

	//Initializing grpc server on listener
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("Cannot listen and start grpc Server: ", err)
	}
}

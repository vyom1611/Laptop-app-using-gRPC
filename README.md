# Laptop-app-using-gRPC
Creating an app to search for laptop configurations using gRPC and protobuf in Go

###### _protoc --proto_path=proto proto/*.proto --go_out=proto --go-grpc_out=proto_


Steps completed till now:
- Created protobuf files
- Created auto generated go files for protobuf
- Added Makefile commands
- Sample generators with protobuf and Go
- Serializing proto messages
- Running end-to-end tests while serializing proto messages to both binary and json
- Comparing temp bins for json and binary files to differentiate between the file sizes, the bin files are more space efficient
- Created a proto service which uses unary-streaming gRPC and implemented a server to handle the unary RPC request
- Added a client to call the unary RPC server and using unit tests for the interaction between server and client
- Added server-streaming RPC to filter laptops while searching in store, then implementing server and client side RPC to handle calls and writing unit tests for them
- Defined client-streaming RPC to upload laptop image to store
- Added client calls and server handlers for client-streaming while implementing unit tests for both
- Implemented Bi-Directional Streaming RPC to create the functionality of rating laptops and writing the unit tests

- Added Evans CLI with gRPC reflection package for more intuitive gRPC actions

## HOW TO RUN THE PROJECT

- download the file and go to its directory in the command line
- run the backend using command `make server`
- open another instance of your command line and run `evans -r repl -p {PORT}` with the port of the backend as the argument
- call `package vyom1611.laptop_app`, and `service LaptopService`
- use the command `show service` to see the available actions and call the desired RPCs using `service {RPC}`
- follow along the prompts showed
- Note: if you want to stop adding details into a list (ex. multiple storages in CreateLaptop), then use `cmd/ctrl + D` to escape
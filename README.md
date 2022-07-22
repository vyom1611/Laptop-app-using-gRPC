# Laptop-app-using-gRPC
Creating an app to search for laptop configurations using gRPC and protobuf in Go

Steps completed till now:
- Created protobuf files
- Created auto generated go files for protobuf
- Added Makefile commands
- Sample generators with protobug and Go
- Serializing proto messages
- Running end-to-end tests while serializing proto messages to both binary and json
- Comparing temp bins for json and binary files to differentiate between the file sizes, the bin files are more space efficient
- Created a proto service which uses unary-streaming gRPC and implemented a server to handle the unary RPC request
- Added a client to call the unary RPC server and using unit tests for the interaction between server and client

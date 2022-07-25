gen: 
	protoc --proto_path=proto proto/*.proto --go_out=plugins=grpc:pb

clean:
	rm pb/pb/*.go

server:
	go run cmd/server/main.go -port 7560

client:
	go run cmd/client/main.go -address 0.0.0.0:7560

test:
	go test -cover -race ./...
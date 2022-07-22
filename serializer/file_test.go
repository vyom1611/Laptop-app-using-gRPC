package serializer_test

import (
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"laptop-app-using-grpc/pb/pb"
	"laptop-app-using-grpc/sample"
	"laptop-app-using-grpc/serializer"
	"testing"
)

func TestFileSerializer(t *testing.T) {
	t.Parallel()

	//Serializing to temp
	binaryFile := "../tmp/laptop.bin"
	jsonFile := "../tmp/laptop.json"

	//Create a laptop
	laptop1 := sample.NewLaptop()
	err := serializer.WriteProtobufToBinaryFile(laptop1, binaryFile)
	require.NoError(t, err)

	laptop2 := &pb.Laptop{}
	err = serializer.ReadProtobufFromBinaryFile(binaryFile, laptop2)
	require.NoError(t, err)
	require.True(t, proto.Equal(laptop1, laptop2))

	err = serializer.WriteProtobufToJSONFile(laptop1, jsonFile)
	require.NoError(t, err)
}

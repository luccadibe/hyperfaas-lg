package internal

import (
	"context"
	"log"

	"github.com/3s-rg-codes/HyperFaaS/proto/common"
	"github.com/3s-rg-codes/HyperFaaS/proto/leaf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type FunctionManager struct {
	conn      *grpc.ClientConn
	client    leaf.LeafClient
	functions map[string]*Function
}

type Function struct {
	ID           string
	ImageTag     string                 `yaml:"image_tag"`
	Timeout      int32                  `yaml:"timeout"`
	DataTemplate map[string]interface{} `yaml:"data_template"`
	ProtoConfig  *common.Config
}

func NewFunctionManager(leafAddress string) *FunctionManager {
	conn, err := grpc.NewClient(leafAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Leaf: %v", err)
	}
	client := leaf.NewLeafClient(conn)
	return &FunctionManager{
		conn:      conn,
		client:    client,
		functions: make(map[string]*Function),
	}
}

func (f *FunctionManager) CreateFunction(imageTag string, timeout int32, dataTemplate map[string]interface{}, protoConfig *common.Config) *Function {

	r, err := f.client.CreateFunction(context.Background(), &leaf.CreateFunctionRequest{
		ImageTag: &common.ImageTag{
			Tag: imageTag,
		},
		Config: protoConfig,
	})
	if err != nil {
		log.Fatalf("Failed to create function: %v", err)
	}

	return &Function{
		ID:           r.FunctionID.Id,
		ImageTag:     imageTag,
		Timeout:      timeout,
		DataTemplate: dataTemplate,
		ProtoConfig:  protoConfig,
	}
}

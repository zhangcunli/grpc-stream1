package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/zhangcunli/grpc-stream1/proto/streammsg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	serverAddr := "localhost:50051"
	grpcStreamTest1(serverAddr)
}

type CustomContext struct {
	ClientInfo ClientInfo `json:"clientInfo"`
	RequestInfo RequestInfo `json:"requestInfo"`
}

type ClientInfo struct {
	ClientId *string `json:"clientId,omitempty"`
	BizType  *int    `json:"bizType,omitempty"`
	AppId    *int    `json:"appId,omitempty"`
}

type RequestInfo struct {
	Lang          string             `json:"lang"`
	NeutralDomain bool               `json:"neutralDomain"`
	BizData       *map[string]string `json:"bizData,omitempty"`
}

func newFusionContext() *CustomContext {
	clientId := "clientId_0001"
	bizType := 10
	clientInfo := ClientInfo{
		ClientId: &clientId,
		BizType:  &bizType,
	}

	bizData := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	requestInfo := RequestInfo{
		Lang:          "zh",
		NeutralDomain: true,
		BizData:       &bizData,
	}
	customContext := CustomContext{
		ClientInfo:  clientInfo,
		RequestInfo: requestInfo,
	}

	return &customContext
}

func grpcStreamTest1(serverAddr string) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	//cp := grpc.ConnectParams{}
	//grpc.WithConnectParams(cp)
	//	grpc.WithChainStreamInterceptor()
	client, err := grpc.NewClient(serverAddr, opts...)
	if err != nil {
		fmt.Printf(">>>grpc NewClient failed, err:%s\n", err)
		return
	}
	defer client.Close()

	client.Connect()
	customContext, _ := json.Marshal(newFusionContext())
	md := metadata.Pairs("customContext", string(customContext))
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	streamDesc := &grpc.StreamDesc{
		ServerStreams: true,
		ClientStreams: true,
	}
	method := "/chatservice.ChatService/Chat"
	stream, err := client.NewStream(ctx, streamDesc, method)
	if err != nil {
		fmt.Printf(">>>grpc NewStream failed, err:%s\n", err)
		return
	}

	waitc := make(chan struct{})
	go func() {
		for {
			resp := new(streammsg.StreamResp)
			err := stream.RecvMsg(resp)
			if err == io.EOF {
				fmt.Printf(">>>stream recv eof\n")
				close(waitc)
				return
			}

			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}
			log.Printf("Got message: %s", resp)
		}
	}()

	index := 0
	go func() {
		for {
			index++
			if index > 20 {
				fmt.Printf(">>>>index:%d close stream\n", index)
				stream.CloseSend()
				return
			}

			msg := &streammsg.StreamReq{
				User: "Client",
				Text: "Hello from the client",
			}
			if err := stream.SendMsg(msg); err != nil {
				log.Fatalf("Failed to send a note: %v", err)
			}

			time.Sleep(1 * time.Second)
		}
	}()

	<-waitc
}


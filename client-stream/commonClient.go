package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/zhangcunli/grpc-stream1/proto/chatservice"
	"github.com/zhangcunli/grpc-stream1/proto/streammsg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//依赖 protoc 生成的 streammsg.pb.go 和 chatservice.pb.go、chatservice_grpc.pb.go
func commonClient(serverAddr string) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	client, err := grpc.NewClient(serverAddr, opts...)
	if err != nil {
		fmt.Printf(">>>grpc NewClient failed, err:%s\n", err)
		return
	}
	defer client.Close()

	client.Connect()
	ctx1, cancel1 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel1()

	c1 := chatservice.NewChatServiceClient(client)
	streamClient, err := c1.Chat(ctx1)
	if err != nil {
		log.Fatalf("Error on get chat: %v", err)
		return
	}

	waitc := make(chan struct{})
	go func() {
		for {
			msg, err := streamClient.Recv()
			if err == io.EOF {
				fmt.Printf(">>>stream recv eof\n")
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}
			log.Printf("Got message %s", msg.GetText())
		}
	}()

	index := 0
	go func() {
		for {
			index++
			if index > 10 {
				streamClient.CloseSend()
				return
			}
			msg := &streammsg.StreamReq{
				User: "Client",
				Text: "Hello from the client",
			}
			if err := streamClient.Send(msg); err != nil {
				log.Fatalf("Failed to send a note: %v", err)
			}
			time.Sleep(1 * time.Second)
		}
	}()
	<-waitc
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

//不依赖任何 protoc 生成的文件，动态生成消息体
func genericStream2(serverAddr string) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	client, err := grpc.NewClient(serverAddr, opts...)
	if err != nil {
		fmt.Printf(">>>grpc NewClient failed, err:%s\n", err)
		return
	}
	defer client.Close()

	client.Connect()
	fusionContext, _ := json.Marshal(newFusionContext())
	md := metadata.Pairs("customContext", string(fusionContext))
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
			//不需要生成 streammsg.pb.go，即可动态生成message结构
			resp := dynamicRespMessage()
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

			//不需要生成 streammsg.pb.go，即可动态生成message结构
			msg := dynamicReqMessage()
			if err := stream.SendMsg(msg); err != nil {
				log.Fatalf("Failed to send a note: %v", err)
			}

			time.Sleep(1 * time.Second)
		}
	}()

	<-waitc
}

//动态生成 streammsg.StreamReq，不依赖 streammsg.pb.go
func dynamicReqMessage() *dynamicpb.Message {
	// 定义一个 proto3 的 message
	msgDesc := &descriptorpb.DescriptorProto{
		Name: proto.String("DynamicMessage"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:   proto.String("user"),
				Number: proto.Int32(1),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
			},
			{
				Name:   proto.String("text"),
				Number: proto.Int32(2),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
			},
			{
				Name:   proto.String("message"),
				Number: proto.Int32(3),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
			},
		},
	}

	// 创建一个 FileDescriptorSet
	fileDescSet := &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{
			{
				Name:    proto.String("dynamic.proto"),
				Package: proto.String("dynamic"),
				Syntax:  proto.String("proto3"),
				MessageType: []*descriptorpb.DescriptorProto{
					msgDesc,
				},
			},
		},
	}

	// 从 FileDescriptorSet 创建一个 FileDescriptorProto
	fileDescProto := fileDescSet.GetFile()[0]

	// 从 FileDescriptorProto 创建一个 FileDescriptor
	fileDesc, err := protodesc.NewFile(fileDescProto, nil)
	if err != nil {
		panic(err)
	}

	// 从 FileDescriptor 获取 MessageDescriptor
	msgDescProto := fileDesc.Messages().Get(0)

	// 创建一个动态的 message
	dynamicMessage := dynamicpb.NewMessage(msgDescProto)

	// 获取字段描述符
	userField := dynamicMessage.Descriptor().Fields().ByName("user")
	textField := dynamicMessage.Descriptor().Fields().ByName("text")

	// 设置字段值
	dynamicMessage.Set(userField, protoreflect.ValueOf("ZhangXX"))
	dynamicMessage.Set(textField, protoreflect.ValueOf("Hello!"))

	// 打印 message
	// fmt.Println(dynamicMessage)

	return dynamicMessage
}

//动态生成 streammsg.StreamResp，不依赖 streammsg.pb.go
func dynamicRespMessage() *dynamicpb.Message {
	// 定义一个 proto3 的 message
	msgDesc := &descriptorpb.DescriptorProto{
		Name: proto.String("DynamicMessage"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:   proto.String("user"),
				Number: proto.Int32(1),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
			},
			{
				Name:   proto.String("text"),
				Number: proto.Int32(2),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
			},
			{
				Name:   proto.String("message"),
				Number: proto.Int32(3),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
			},
		},
	}

	// 创建一个 FileDescriptorSet
	fileDescSet := &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{
			{
				Name:    proto.String("dynamic.proto"),
				Package: proto.String("dynamic"),
				Syntax:  proto.String("proto3"),
				MessageType: []*descriptorpb.DescriptorProto{
					msgDesc,
				},
			},
		},
	}

	// 从 FileDescriptorSet 创建一个 FileDescriptorProto
	fileDescProto := fileDescSet.GetFile()[0]

	// 从 FileDescriptorProto 创建一个 FileDescriptor
	fileDesc, err := protodesc.NewFile(fileDescProto, nil)
	if err != nil {
		panic(err)
	}

	// 从 FileDescriptor 获取 MessageDescriptor
	msgDescProto := fileDesc.Messages().Get(0)

	// 创建一个动态的 message
	dynamicMessage := dynamicpb.NewMessage(msgDescProto)

	return dynamicMessage
}

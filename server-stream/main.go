package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"strings"

	"github.com/zhangcunli/grpc-stream1/proto/chatservice"
	"github.com/zhangcunli/grpc-stream1/proto/streammsg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type chatServer struct {
	chatservice.UnimplementedChatServiceServer
}

func (s *chatServer) Chat(stream chatservice.ChatService_ChatServer) error {
	// 从 context 中读取元数据
	md, ok := metadata.FromIncomingContext(stream.Context())
	if ok {
		k := strings.ToLower("fusionContext")
		values, exist := md[k]
		if exist {
			value := values[0]
			log.Printf("Chat in metadata, key=%s, value=%s, value type:%v", k, value, reflect.TypeOf(values))
		} else {
			fmt.Printf(">>>>chat in metadata key:%s is null", k)
		}
	}

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			fmt.Printf(">>>Chat message EOF\n")
			return nil
		}
		if err != nil {
			fmt.Printf(">>>Chat message err:%s\n", err.Error())
			return err
		}
		fmt.Printf(">>>Chat recv msg:%s\n", msg)
		stream.Send(&streammsg.StreamResp{
			User: "Server",
			Text: "Hello " + msg.GetUser(),
		})
	}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	//kp := keepalive.ServerParameters{}
	//grpc.KeepaliveParams(kp)
	//kep := keepalive.EnforcementPolicy{}
	//grpc.KeepaliveEnforcementPolicy(kep)
	s := grpc.NewServer()
	chatservice.RegisterChatServiceServer(s, &chatServer{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

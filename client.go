// +build ignore

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"

	"github.com/navigaid/grpc-boomerang/pkg/api"
)

var addr = flag.String("addr", "localhost:8080", "tcp service address")

/*
[client]

tcpSide <= io.Pipe =>  grpcSide = (grpc.Serve)

(raw data)

^ goroutine 1
v goroutine 2

(binary message)

|| websocket

[server]

*/

func main() {
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Printf("connecting to %s", *addr)

	c, err := net.Dial("tcp", *addr)
	if err != nil {
		log.Fatal("dial:", err)
	}

	grpcHandler := &apiService{}

	// client side grpc server over net.Conn over websocket.Conn
	l := &singleListener{c}
	grpcServer := grpc.NewServer()
	api.RegisterApiServer(grpcServer, grpcHandler)
	grpcServer.Serve(l)
}

// single listener converts/upgrades the current tcp connection into grpc
// ============================= gender changer impl
type singleListener struct {
	conn net.Conn
}

func (s *singleListener) Accept() (net.Conn, error) {
	if s.conn != nil {
		log.Println("Gender Change: TCP Client -> GRPC Server")
		c := s.conn
		s.conn = nil
		return c, nil
	}
	select {}
	return nil, nil
}

func (s *singleListener) Close() error {
	return nil
}

func (s *singleListener) Addr() net.Addr {
	return s.conn.LocalAddr()
}

// apiService acts as the real grpc request handler
// ============================= api impl
type apiService struct {
}

func (s *apiService) Hello(ctx context.Context, in *api.HelloRequest) (*api.HelloResponse, error) {
	return &api.HelloResponse{Message: "Hello " + in.GetName()}, nil
}

func (s *apiService) HelloStream(in *api.HelloStreamRequest, stream api.Api_HelloStreamServer) error {
	for i := 0; i < 10; i++ {
		err := stream.Send(&api.HelloStreamResponse{Message: fmt.Sprintf("Hello %d: %s", i, in.GetName())})
		time.Sleep(0 * time.Second)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *apiService) StdinStream(in *api.StdinStreamRequest, stream api.Api_StdinStreamServer) error {
	buf := make([]byte, 1<<16)
	file, err := os.Open(in.Name)
	if err != nil {
		return err
	}
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		err = stream.Send(&api.StdinStreamResponse{Message: buf[:n]})
		if err != nil {
			return err
		}
	}
	return nil
}

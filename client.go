// +build ignore

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc"

	"github.com/navigaid/grpc-boomerang/pkg/api"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

/*
[client]

websocketSide  <= io.Pipe =>  grpcSide = (grpc.Serve)

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

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	grpcSide, websocketSide := net.Pipe()

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})
	data := make([]byte, 10*1024*1024)

	go func() {
		defer c.Close()
		defer close(done)
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("c.ReadMessage:", err)
				break
			}

			if mt != websocket.BinaryMessage {
				log.Println("mt != websocket.BinaryMessage")
				break
			}

			n, err := websocketSide.Write(message)
			if err != nil {
				log.Println("pipe.Write:", err)
				break
			}

			if len(message) != n {
				log.Printf("whooot! len(data) != n => %d != %d!\n", len(message), n)
				break
			}
		}
	}()

	go func() {
		for {
			n, err := websocketSide.Read(data)
			if err != nil {
				log.Println("pipe.Read:", err)
				break
			}

			err = c.WriteMessage(websocket.BinaryMessage, data[:n])
			if err != nil {
				log.Println("c.WriteMessage:", err)
				break
			}
		}
	}()

	// client side grpc server over net.Conn over websocket.Conn
	l := &singleListener{grpcSide}
	grpcServer := grpc.NewServer()
	api.RegisterApiServer(grpcServer, &apiService{})
	err = grpcServer.Serve(l)
	if err != nil {
		log.Println("grpcServer.Serve", err)
	}
	select {}
}

type singleListener struct {
	conn net.Conn
}

func (s *singleListener) Accept() (net.Conn, error) {
	if s.conn != nil {
		log.Println("Accept")
		c := s.conn
		s.conn = nil
		return c, nil
	}
	log.Println("Reject")
	return nil, io.EOF
}

func (s *singleListener) Close() error {
	return nil
}

func (s *singleListener) Addr() net.Addr {
	return s.conn.LocalAddr()
}

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

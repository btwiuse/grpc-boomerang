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
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc"

	"github.com/navigaid/grpc-boomerang/pkg/api"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

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
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
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

	wg.Add(1)
	go func() {
		defer wg.Done()
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

	l := &singleListener{grpcSide}
	grpcServer := grpc.NewServer()
	api.RegisterApiServer(grpcServer, &apiService{})
	grpcServer.Serve(l)

	for {
		select {
		/*
			case t := <-ticker.C:
				err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
				if err != nil {
					log.Println("write:", err)
					return
				}
		*/
		case <-interrupt:
			log.Println("interrupt")
			// To cleanly close a connection, a client should send a close
			// frame and wait for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
				wg.Wait()
			case <-time.After(time.Second):
			}
			c.Close()
			return
		}
	}
}

type singleListener struct {
	conn net.Conn
}

func (s *singleListener) Accept() (net.Conn, error) {
	log.Println("Accept")
	if c := s.conn; c != nil {
		s.conn = nil
		return c, nil
	}
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
		time.Sleep(1 * time.Second)
		if err != nil {
			return err
		}
	}
	return nil
}

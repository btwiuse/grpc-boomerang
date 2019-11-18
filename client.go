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
	"google.golang.org/grpc/credentials"

	"github.com/btwiuse/grpc-boomerang/pkg/api"
	"github.com/btwiuse/wetty/localcmd"
)

var addr = flag.String("addr", "localhost:8443", "tcp service address")
var cancel func()
var ctx context.Context

func main() {
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	c, err := net.Dial("tcp", *addr)
	if err != nil {
		log.Fatal("dial:", err)
	}

	log.Printf("connecting to %s from %s\n", *addr, c.LocalAddr())

	creds, err := credentials.NewServerTLSFromFile("localhost.pem", "localhost-key.pem")
	if err != nil {
		log.Fatalln("bad credentials:", err)
	}

	options := []grpc.ServerOption{
		grpc.Creds(creds),
	}
	grpcServer := grpc.NewServer(options...)
	api.RegisterApiServer(grpcServer, &apiService{})
	grpcServer.Serve(&singleListener{pipe(c)})
}

func pipe(c net.Conn) (net.Conn){
	errs := make(chan error, 2)
	a, b := net.Pipe()

	go func() {
		defer b.Close()
		defer c.Close()
		_, err := io.Copy(b, c)
		errs <- err
	}()

	go func() {
		defer b.Close()
		defer c.Close()
		_, err := io.Copy(c, b)
		errs <- err
	}()

	ctx, cancel = context.WithCancel(context.Background())
	go func() {
		<-errs
		cancel()
	}()

	return a
}

// single listener converts/upgrades the current tcp connection into grpc
// ============================= gender changer impl
type singleListener struct {
	net.Conn
}

func (s *singleListener) Accept() (net.Conn, error) {
	if s.Conn != nil {
		log.Println("Gender Change: TCP Client -> GRPC Server")
		c := s.Conn
		s.Conn = nil
		return c, nil
	}
	<-ctx.Done()
	os.Exit(1)
	return nil, nil
}

func (s *singleListener) Close() error {
	return nil
}

func (s *singleListener) Addr() net.Addr {
	return s.Conn.LocalAddr()
}

// apiService acts as the real grpc request handler
// ============================= api impl
type apiService struct{}

func (s *apiService) Probe(ctx context.Context, ping *api.Ping) (*api.Pong, error) {
	log.Println("Ping received. Sending Pong.")
	return &api.Pong{}, nil
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

func (s *apiService) HtopStream(in *api.HtopStreamRequest, stream api.Api_HtopStreamServer) error {
	lc, err := localcmd.NewLc([]string{"htop"})
	if err != nil {
		return err
	}

	buf := make([]byte, 1<<16)
	// file, err := os.Open(in.Name)
	if err != nil {
		return err
	}
	for {
		// n, err := file.Read(buf)
		n, err := lc.Read(buf)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		err = stream.Send(&api.HtopStreamResponse{Message: buf[:n]})
		if err != nil {
			return err
		}
	}
	return nil
}

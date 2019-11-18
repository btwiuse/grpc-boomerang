// +build ignore

package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/btwiuse/grpc-boomerang/pkg/api"
	"github.com/btwiuse/grpc-boomerang/pkg/api/impl"
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
	api.RegisterApiServer(grpcServer, &impl.ApiService{})
	grpcServer.Serve(&singleListener{pipe(c)})
}

func pipe(c net.Conn) net.Conn {
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

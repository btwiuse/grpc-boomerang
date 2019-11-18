// +build ignore

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/btwiuse/grpc-boomerang/pkg/api"
)

var (
	addr = flag.String("addr", "localhost:8443", "http service address")
	creds credentials.TransportCredentials
)

func main() {
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("listening on", *addr)
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalln(err)
	}

	creds, err = credentials.NewClientTLSFromFile("localhost.pem", "localhost")
	if err != nil {
		log.Fatalln("bad credentials:", err)
	}

	for {
		c, err := ln.Accept()
		if err != nil {
			continue
		}

		go handle(c)
	}
}

func pipe(c net.Conn) (context.Context, net.Conn) {
	errs := make(chan error, 2)
	a, b := net.Pipe()

	go func() {
		defer c.Close()
		defer b.Close()
		_, err := io.Copy(b, c)
		errs <- err
	}()

	go func() {
		defer c.Close()
		defer b.Close()
		_, err := io.Copy(c, b)
		errs <- err
	}()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-errs
		cancel()
	}()
	return ctx, a
}

func toClientConn(c net.Conn) *grpc.ClientConn {
	options := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithContextDialer(
			func(ctx context.Context, s string) (net.Conn, error) {
				return c, nil
			},
		),
	}

	cc, err := grpc.Dial("", options...)
	if err != nil {
		log.Fatal(err.Error())
	}
	return cc
}

func handle(c net.Conn) {
	log.Println("new client:", c.RemoteAddr())
	ctx, a := pipe(c)
	client := api.NewApiClient(toClientConn(a))

	streamRequest := &api.HtopStreamRequest{}
	log.Println("sending HtopStreamRequest")
	streamResponse, err := client.HtopStream(context.Background(), streamRequest)
	if err != nil {
		log.Fatalf("%v.HtopStream(_) = _, %v", client, err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("client dead:", c.RemoteAddr())
			return
		default:
			streamResponseItem, err := streamResponse.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Println(err)
				return
			}
			fmt.Printf("%s", streamResponseItem.Message)
		}
	}
}

/*
{
	streamRequest := &api.StdinStreamRequest{Name: string(l)}
	log.Println("sending StdinStreamRequest")
	streamResponse, err := client.StdinStream(context.Background(), streamRequest)
	if err != nil {
		log.Fatalf("%v.StdinStream(_) = _, %v", client, err)
	}

	for i := 0; ; i++ {
		streamResponseItem, err := streamResponse.Recv()
		if err != nil {
			log.Printf("streamResponse.Recv() %v\n", err)
			break
		}
		log.Println(i, "StdinStreamResponse", len(streamResponseItem.Message))
	}
}
{
	streamRequest := &api.HelloStreamRequest{Name: string(l)}
	log.Println("sending HelloStreamRequest")
	streamResponse, err := client.HelloStream(context.Background(), streamRequest)
	if err != nil {
		log.Fatalf("%v.HelloStream(_) = _, %v", client, err)
	}

	for {
		streamResponseItem, err := streamResponse.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("streamResponse.Recv() %v", err)
		}
		log.Println("HelloStreamResponse", streamResponseItem)
	}
}
{
	request := &api.HelloRequest{Name: string(l)}
	log.Println("sending HelloRequest")
	response, err := client.Hello(context.Background(), request)
	if err != nil {
		log.Println(err.Error())
		continue
	}
	log.Println("HelloResponse" + response.GetMessage())
}
*/

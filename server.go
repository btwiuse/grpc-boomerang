// +build ignore

package main

import (
	"context"
	"flag"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/navigaid/grpc-boomerang/pkg/api"
)

var (
	addr = flag.String("addr", "localhost:8080", "http service address")
)

func main() {
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("listening on", *addr)
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalln(err)
	}
	for {
		c, err := ln.Accept()
		if err != nil {
			continue
		}
		cc := convert(c)
		client := api.NewApiClient(cc)
		go handle(client)
	}
}

func convert(c net.Conn) *grpc.ClientConn {
	log.Println("net.Conn -> grpc client")
	cc, err := grpc.Dial("",
		grpc.WithInsecure(),
		grpc.WithContextDialer(
			func(ctx context.Context, s string) (net.Conn, error) {
				return c, nil
			},
		),
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	return cc
}

func handle(client api.ApiClient) {
	for range time.Tick(time.Second) {
		client.Probe(context.Background(), &api.Ping{})
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

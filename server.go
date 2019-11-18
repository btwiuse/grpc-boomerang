// +build ignore

package main

import (
	"bufio"
	"context"
	"flag"
	"io"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/navigaid/grpc-boomerang/pkg/api"
)

var (
	addr                    = flag.String("addr", "localhost:8080", "http service address")
	grpcSide, websocketSide = net.Pipe()
)

func echo(grpcSide, websocketSide net.Conn) func(net.Conn) {
	return func(c net.Conn) {
		go func() {
			defer c.Close()
			log.Println(io.Copy(websocketSide, c))
		}()

		go func() {
			defer c.Close()
			log.Println(io.Copy(c, websocketSide))
		}()

		select {}
	}

}

func main() {
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	go func() {
		log.Println("listening on", *addr)
		ln, err := net.Listen("tcp", *addr)
		if err != nil {
			log.Fatalln(err)
		}
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		echo(grpcSide, websocketSide)(conn)
	}()

	log.Println("grpc.Dial")
	c, err := grpc.Dial("",
		grpc.WithInsecure(),
		grpc.WithContextDialer(
			func(ctx context.Context, s string) (net.Conn, error) {
				return grpcSide, nil
			},
		),
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	client := api.NewApiClient(c)

	r := bufio.NewReader(os.Stdin)
	for {
		l, _, err := r.ReadLine()
		if err != nil {
			log.Fatal(err.Error())
		}

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
			//streamRequest := &api.HelloStreamRequest{Name: string(l)}
			//log.Println("sending HelloStreamRequest")
			//streamResponse, err := client.HelloStream(context.Background(), streamRequest)
			//if err != nil {
			//	log.Fatalf("%v.HelloStream(_) = _, %v", client, err)
			//}

			//for {
			//	streamResponseItem, err := streamResponse.Recv()
			//	if err == io.EOF {
			//		break
			//	}
			//	if err != nil {
			//		log.Fatalf("streamResponse.Recv() %v", err)
			//	}
			//	log.Println("HelloStreamResponse", streamResponseItem)
			//}
		}
		{
			//request := &api.HelloRequest{Name: string(l)}
			//log.Println("sending HelloRequest")
			//response, err := client.Hello(context.Background(), request)
			//if err != nil {
			//	log.Println(err.Error())
			//	continue
			//}
			//log.Println("HelloResponse" + response.GetMessage())
		}
	}
}

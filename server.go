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

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/btwiuse/grpc-boomerang/pkg/api"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	addr  = flag.String("addr", "localhost:8443", "http service address")
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
	oldState, _ := terminal.MakeRaw(0)
	if err != nil {
		log.Fatalln(err)
	}
	defer func(){
		terminal.Restore(0, oldState)
	}()

	log.Println("new client:", c.RemoteAddr())
	ctx, a := pipe(c)
	cc := toClientConn(a)
	bidiStreamClient := api.NewBidiStreamClient(cc)

	log.Println("new bidiStreamClient:", bidiStreamClient)

	sendClient, err := bidiStreamClient.Send(ctx)
	if err != nil {
		log.Fatalln(bidiStreamClient, err)
	}

	log.Println("new sendClient:", sendClient)

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				log.Println(err)
				break
			}

			inputMsg := &api.Message{Type: []byte{2}, Body: buf[:n]}
			err = sendClient.Send(inputMsg)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("client dead:", c.RemoteAddr())
			return
		default:
			resp, err := sendClient.Recv()
			if err != nil {
				log.Println(err)
				return
			}
			fmt.Printf("%s", resp.Body)
		}
	}
}

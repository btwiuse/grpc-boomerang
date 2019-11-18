// +build ignore

// client:
// docker run -it -v $PWD/index:/index -v $PWD/log:/log --env-file=config .env --net=host navigaid/grpcox

package main

import (
	"flag"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/btwiuse/grpc-boomerang/pkg/api"
	"github.com/btwiuse/grpc-boomerang/pkg/api/impl"
	"google.golang.org/grpc/reflection"
)

var (
	addr  = flag.String("addr", "localhost:8080", "http service address")
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

	/*
		creds, err = credentials.NewServerTLSFromFile("localhost.pem", "localhost-key.pem")
		if err != nil {
			log.Fatalln("bad credentials:", err)
		}
	*/

	options := []grpc.ServerOption{
		// grpc.Creds(creds),
	}
	grpcServer := grpc.NewServer(options...)
	reflection.Register(grpcServer)
	api.RegisterApiServer(grpcServer, &impl.ApiService{})

	grpcServer.Serve(ln)
}

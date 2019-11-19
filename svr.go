// +build ignore

/*
# https://github.com/grpc/grpc-go/blob/master/Documentation/server-reflection-tutorial.md

web-ui client:
$ docker run -it -v $PWD/index:/index -v $PWD/log:/log --env-file=config .env --net=host navigaid/grpcox

command-line client:
$ grpc_cli ls localhost:8080
$ grpc_cli ls localhost:8080 Api.Probe -l
  rpc Probe(Ping) returns (Pong) {}
$ grpc_cli ls localhost:8080 Api.HtopStream -l
  rpc HtopStream(HtopStreamRequest) returns (stream HtopStreamResponse) {}
$ </dev/null grpc_cli call localhost:8080 Api.HtopStream
  ...
$ </dev/null grpc_cli call localhost:8080 Api.Probe
connecting to localhost:8080

Rpc succeeded with OK status

*/

package main

import (
	"flag"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	"github.com/btwiuse/grpc-boomerang/pkg/api"
	"github.com/btwiuse/grpc-boomerang/pkg/api/impl"
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

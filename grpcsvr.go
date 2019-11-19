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
	"fmt"
	"log"
	// "net"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"grpc.go4.org/grpclog"

	"github.com/btwiuse/grpc-boomerang/pkg/api"
	"github.com/btwiuse/grpc-boomerang/pkg/api/impl"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
)

var (
	addr  = flag.String("addr", "localhost:9090", "http service address")
	creds credentials.TransportCredentials
)

func main() {
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("listening on", *addr)
	/*
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalln(err)
	}

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

	wrappedServer := grpcweb.WrapServer(grpcServer,
		grpcweb.WithAllowedRequestHeaders([]string{
			"x-grpc-web",
		}),
	)

	http.Handle("/", http.FileServer(http.Dir("ts")))
	http.Handle("/Api/", wrappedServer)

	httpServer := http.Server{
		Addr:    fmt.Sprintf(":%d", 9090),
		// Handler: cors.Default().Handler(http.HandlerFunc(handler)),
		Handler: http.DefaultServeMux,
	}

	// _ = ln
	// grpcServer.Serve(ln)
	if err := httpServer.ListenAndServe(); err != nil {
		grpclog.Fatalf("failed starting http server: %v", err)
	}
}

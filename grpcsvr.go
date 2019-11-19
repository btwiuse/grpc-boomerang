// +build ignore

package main

import (
	"flag"
	"log"
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

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	api.RegisterApiServer(grpcServer, &impl.ApiService{})

	wrappedServer := grpcweb.WrapServer(grpcServer)

	http.Handle("/", http.FileServer(http.Dir("ts")))
	http.Handle("/Api/", wrappedServer)

	httpServer := http.Server{
		Addr:    *addr,
		Handler: http.DefaultServeMux,
	}

	if err := httpServer.ListenAndServe(); err != nil {
		grpclog.Fatalf("failed starting http server: %v", err)
	}
}

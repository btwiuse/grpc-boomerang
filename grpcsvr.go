// +build ignore

package main

import (
	"flag"
	"log"
	"net/http"

	"google.golang.org/grpc"
	"grpc.go4.org/grpclog"

	"github.com/btwiuse/grpc-boomerang/pkg/api"
	"github.com/btwiuse/grpc-boomerang/pkg/api/impl"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
)

var (
	port            = flag.String("port", ":9090", "http service address")
	enableTls       = flag.Bool("enable_tls", false, "Use TLS - required for HTTP2.")
	tlsCertFilePath = flag.String("tls_cert_file", "localhost.pem", "Path to the CRT/PEM file.")
	tlsKeyFilePath  = flag.String("tls_key_file", "localhost-key.pem", "Path to the private key file.")
)

func main() {
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	grpcServer := grpc.NewServer()
	api.RegisterApiServer(grpcServer, &impl.ApiService{})

	wrappedServer := grpcweb.WrapServer(grpcServer)
	fileServer := http.FileServer(http.Dir("ts"))

	http.Handle("/Api/", wrappedServer)
	http.Handle("/", fileServer)

	httpServer := http.Server{
		Addr:    *port,
		Handler: http.DefaultServeMux,
	}

	if *enableTls {
		grpclog.Printf("Starting server. https://localhost%s", *port)
		if err := httpServer.ListenAndServeTLS(*tlsCertFilePath, *tlsKeyFilePath); err != nil {
			grpclog.Fatalf("failed starting http2 server: %v", err)
		}
	} else {
		grpclog.Printf("Starting server. http://localhost%s", *port)
		if err := httpServer.ListenAndServe(); err != nil {
			grpclog.Fatalf("failed starting http server: %v", err)
		}
	}
}

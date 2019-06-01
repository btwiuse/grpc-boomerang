// +build ignore

package main

import (
	"bufio"
	"context"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc"

	"github.com/navigaid/grpc-boomerang/pkg/api"
)

var (
	addr                    = flag.String("addr", "localhost:8080", "http service address")
	upgrader                = websocket.Upgrader{} // use default options
	mux                     = http.NewServeMux()
	grpcSide, websocketSide = net.Pipe()
)

func echo(grpcSide, websocketSide net.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()

		data := make([]byte, 10*1024*1024)
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				mt, message, err := c.ReadMessage()
				if err != nil {
					log.Println("c.ReadMessage:", err)
					break
				}

				if mt != websocket.BinaryMessage {
					log.Println("mt != websocket.BinaryMessage")
					break
				}

				n, err := websocketSide.Write(message)
				if err != nil {
					log.Println("pipe.Write:", err)
					break
				}

				if len(message) != n {
					log.Printf("whooot! len(data) != n => %d != %d!\n", len(message), n)
					break
				}
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				n, err := websocketSide.Read(data)
				if err != nil {
					log.Println("pipe.Read:", err)
					break
				}

				err = c.WriteMessage(websocket.BinaryMessage, data[:n])
				if err != nil {
					log.Println("c.WriteMessage:", err)
					break
				}
			}
		}()

		wg.Wait()
	}

}

func main() {
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	http.HandleFunc("/echo", echo(grpcSide, websocketSide))
	go func() {
		log.Println("listening on", *addr)
		log.Fatalln(http.ListenAndServe(*addr, nil))
	}()

	c, err := grpc.Dial("", []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return grpcSide, nil
		}),
	}...,
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

		go func() {
			streamRequest := &api.HelloStreamRequest{Name: string(l)}
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
				log.Println(streamResponseItem)
			}
		}()

		println("hello")
		request := &api.HelloRequest{Name: string(l)}
		response, err := client.Hello(context.Background(), request)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		println("response: " + response.GetMessage())
	}
}

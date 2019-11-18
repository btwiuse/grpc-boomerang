// +build ignore

package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"

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

		go func() {
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

				_, err = websocketSide.Write(message)
				if err != nil {
					log.Println("pipe.Write:", err)
					break
				}
			}
		}()

		go func() {
			data := make([]byte, 10*1024*1024)
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

		select {}
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

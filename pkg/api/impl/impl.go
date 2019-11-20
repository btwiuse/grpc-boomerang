package impl

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/btwiuse/grpc-boomerang/pkg/api"
	"github.com/btwiuse/wetty/localcmd"
)

type BidiStream struct{}

func (bs *BidiStream) Send(sendServer api.BidiStream_SendServer) error {
	// send
	go func(){
                emptyMsg := &api.Message{Type: []byte{1}, Body: []byte{}}
		for range time.Tick(time.Second) {
                        log.Println("sending empty message", emptyMsg.Type, emptyMsg.Body)
                        err := sendServer.Send(emptyMsg)
                        if err != nil {
                                log.Println(err)
                                break
                        }
		}
	}()

	// recv
	for {
		msg, err := sendServer.Recv()
		if err != nil {
			return nil
		}
		log.Println(msg.Type, msg.Body)
	}

	return nil
}

// apiService acts as the real grpc request handler
// ============================= api impl
type ApiService struct{}

func (s *ApiService) Probe(ctx context.Context, ping *api.Ping) (*api.Pong, error) {
	log.Println("Ping received. Sending Pong.")
	return &api.Pong{}, nil
}

func (s *ApiService) Hello(ctx context.Context, in *api.HelloRequest) (*api.HelloResponse, error) {
	return &api.HelloResponse{Message: "Hello " + in.GetName()}, nil
}

func (s *ApiService) HelloStream(in *api.HelloStreamRequest, stream api.Api_HelloStreamServer) error {
	for i := 0; i < 10; i++ {
		err := stream.Send(&api.HelloStreamResponse{Message: fmt.Sprintf("Hello %d: %s", i, in.GetName())})
		time.Sleep(0 * time.Second)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ApiService) StdinStream(in *api.StdinStreamRequest, stream api.Api_StdinStreamServer) error {
	buf := make([]byte, 1<<16)
	file, err := os.Open(in.Name)
	if err != nil {
		return err
	}
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		err = stream.Send(&api.StdinStreamResponse{Message: buf[:n]})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ApiService) HtopStream(in *api.HtopStreamRequest, stream api.Api_HtopStreamServer) error {
	lc, err := localcmd.NewLc([]string{"htop"})
	if err != nil {
		return err
	}

	buf := make([]byte, 1<<16)
	// file, err := os.Open(in.Name)
	if err != nil {
		return err
	}
	for {
		// n, err := file.Read(buf)
		n, err := lc.Read(buf)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		err = stream.Send(&api.HtopStreamResponse{Message: buf[:n]})
		if err != nil {
			return err
		}
	}
	return nil
}

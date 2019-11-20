import {grpc} from "@improbable-eng/grpc-web";
import {Api} from "../../pkg/api/api_pb_service";
import {Ping, HelloRequest, HelloResponse, HelloStreamRequest, HelloStreamResponse , HtopStreamRequest, HtopStreamResponse} from "../../pkg/api/api_pb";

const host = "//localhost:9090";

function hello() {
  const helloRequest = new HelloRequest();
  helloRequest.setName("Joe");
  grpc.unary(Api.Hello, {
    request: helloRequest,
    host: host,
    onEnd: res => {
      const { status, statusMessage, headers, message, trailers } = res;
      console.log("hello.onEnd.status", status, statusMessage);
      console.log("hello.onEnd.headers", headers);
      if (status === grpc.Code.OK && message) {
        console.log("hello.onEnd.message", message.toObject());
      }
      console.log("hello.onEnd.trailers", trailers);
    }
  });
}

function helloStream() {
  const helloStreamRequest = new HelloStreamRequest();
  helloStreamRequest.setName("Geor");
  const client = grpc.client(Api.HelloStream, {
    host: host,
  });
  client.onMessage((message: HelloStreamResponse) => {
    console.log(message.toObject().message);
  });
  client.start();
  client.send(helloStreamRequest);
}

function htopStream() {
  const htopStreamRequest = new HtopStreamRequest();
  const client = grpc.client(Api.HtopStream, {
    host: host,
  });
  client.onMessage((message: HtopStreamResponse) => {
    console.log(message.toObject().message);
  });
  client.start();
  client.send(htopStreamRequest);
}

// hello();
// helloStream();
htopStream();

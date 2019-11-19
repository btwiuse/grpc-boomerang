import {grpc} from "@improbable-eng/grpc-web";
import {Api} from "../../pkg/api/api_pb_service";
import {Ping, HelloRequest, HelloResponse, HelloStreamRequest, HelloStreamResponse } from "../../pkg/api/api_pb";

declare const USE_TLS: boolean;
const host = USE_TLS ? "https://localhost:9091" : "http://localhost:9090";

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
  client.onHeaders((headers: grpc.Metadata) => {
    console.log("helloStream.onHeaders", headers);
  });
  client.onMessage((message: HelloStreamResponse) => {
    console.log("helloStream.onMessage", message.toObject());
  });
  client.onEnd((code: grpc.Code, msg: string, trailers: grpc.Metadata) => {
    console.log("helloStream.onEnd", code, msg, trailers);
  });
  client.start();
  client.send(helloStreamRequest);
}

hello();
helloStream();

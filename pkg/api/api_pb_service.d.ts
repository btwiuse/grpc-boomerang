// package: 
// file: api.proto

import * as api_pb from "./api_pb";
import {grpc} from "@improbable-eng/grpc-web";

type ApiProbe = {
  readonly methodName: string;
  readonly service: typeof Api;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof api_pb.Ping;
  readonly responseType: typeof api_pb.Pong;
};

type ApiHello = {
  readonly methodName: string;
  readonly service: typeof Api;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof api_pb.HelloRequest;
  readonly responseType: typeof api_pb.HelloResponse;
};

type ApiHelloStream = {
  readonly methodName: string;
  readonly service: typeof Api;
  readonly requestStream: false;
  readonly responseStream: true;
  readonly requestType: typeof api_pb.HelloStreamRequest;
  readonly responseType: typeof api_pb.HelloStreamResponse;
};

type ApiStdinStream = {
  readonly methodName: string;
  readonly service: typeof Api;
  readonly requestStream: false;
  readonly responseStream: true;
  readonly requestType: typeof api_pb.StdinStreamRequest;
  readonly responseType: typeof api_pb.StdinStreamResponse;
};

type ApiHtopStream = {
  readonly methodName: string;
  readonly service: typeof Api;
  readonly requestStream: false;
  readonly responseStream: true;
  readonly requestType: typeof api_pb.HtopStreamRequest;
  readonly responseType: typeof api_pb.HtopStreamResponse;
};

export class Api {
  static readonly serviceName: string;
  static readonly Probe: ApiProbe;
  static readonly Hello: ApiHello;
  static readonly HelloStream: ApiHelloStream;
  static readonly StdinStream: ApiStdinStream;
  static readonly HtopStream: ApiHtopStream;
}

type BidiStreamSend = {
  readonly methodName: string;
  readonly service: typeof BidiStream;
  readonly requestStream: true;
  readonly responseStream: true;
  readonly requestType: typeof api_pb.Message;
  readonly responseType: typeof api_pb.Message;
};

export class BidiStream {
  static readonly serviceName: string;
  static readonly Send: BidiStreamSend;
}

export type ServiceError = { message: string, code: number; metadata: grpc.Metadata }
export type Status = { details: string, code: number; metadata: grpc.Metadata }

interface UnaryResponse {
  cancel(): void;
}
interface ResponseStream<T> {
  cancel(): void;
  on(type: 'data', handler: (message: T) => void): ResponseStream<T>;
  on(type: 'end', handler: (status?: Status) => void): ResponseStream<T>;
  on(type: 'status', handler: (status: Status) => void): ResponseStream<T>;
}
interface RequestStream<T> {
  write(message: T): RequestStream<T>;
  end(): void;
  cancel(): void;
  on(type: 'end', handler: (status?: Status) => void): RequestStream<T>;
  on(type: 'status', handler: (status: Status) => void): RequestStream<T>;
}
interface BidirectionalStream<ReqT, ResT> {
  write(message: ReqT): BidirectionalStream<ReqT, ResT>;
  end(): void;
  cancel(): void;
  on(type: 'data', handler: (message: ResT) => void): BidirectionalStream<ReqT, ResT>;
  on(type: 'end', handler: (status?: Status) => void): BidirectionalStream<ReqT, ResT>;
  on(type: 'status', handler: (status: Status) => void): BidirectionalStream<ReqT, ResT>;
}

export class ApiClient {
  readonly serviceHost: string;

  constructor(serviceHost: string, options?: grpc.RpcOptions);
  probe(
    requestMessage: api_pb.Ping,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: api_pb.Pong|null) => void
  ): UnaryResponse;
  probe(
    requestMessage: api_pb.Ping,
    callback: (error: ServiceError|null, responseMessage: api_pb.Pong|null) => void
  ): UnaryResponse;
  hello(
    requestMessage: api_pb.HelloRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: api_pb.HelloResponse|null) => void
  ): UnaryResponse;
  hello(
    requestMessage: api_pb.HelloRequest,
    callback: (error: ServiceError|null, responseMessage: api_pb.HelloResponse|null) => void
  ): UnaryResponse;
  helloStream(requestMessage: api_pb.HelloStreamRequest, metadata?: grpc.Metadata): ResponseStream<api_pb.HelloStreamResponse>;
  stdinStream(requestMessage: api_pb.StdinStreamRequest, metadata?: grpc.Metadata): ResponseStream<api_pb.StdinStreamResponse>;
  htopStream(requestMessage: api_pb.HtopStreamRequest, metadata?: grpc.Metadata): ResponseStream<api_pb.HtopStreamResponse>;
}

export class BidiStreamClient {
  readonly serviceHost: string;

  constructor(serviceHost: string, options?: grpc.RpcOptions);
  send(metadata?: grpc.Metadata): BidirectionalStream<api_pb.Message, api_pb.Message>;
}


// package: 
// file: api.proto

var api_pb = require("./api_pb");
var grpc = require("@improbable-eng/grpc-web").grpc;

var Api = (function () {
  function Api() {}
  Api.serviceName = "Api";
  return Api;
}());

Api.Probe = {
  methodName: "Probe",
  service: Api,
  requestStream: false,
  responseStream: false,
  requestType: api_pb.Ping,
  responseType: api_pb.Pong
};

Api.Hello = {
  methodName: "Hello",
  service: Api,
  requestStream: false,
  responseStream: false,
  requestType: api_pb.HelloRequest,
  responseType: api_pb.HelloResponse
};

Api.HelloStream = {
  methodName: "HelloStream",
  service: Api,
  requestStream: false,
  responseStream: true,
  requestType: api_pb.HelloStreamRequest,
  responseType: api_pb.HelloStreamResponse
};

Api.StdinStream = {
  methodName: "StdinStream",
  service: Api,
  requestStream: false,
  responseStream: true,
  requestType: api_pb.StdinStreamRequest,
  responseType: api_pb.StdinStreamResponse
};

Api.HtopStream = {
  methodName: "HtopStream",
  service: Api,
  requestStream: false,
  responseStream: true,
  requestType: api_pb.HtopStreamRequest,
  responseType: api_pb.HtopStreamResponse
};

exports.Api = Api;

function ApiClient(serviceHost, options) {
  this.serviceHost = serviceHost;
  this.options = options || {};
}

ApiClient.prototype.probe = function probe(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(Api.Probe, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

ApiClient.prototype.hello = function hello(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(Api.Hello, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

ApiClient.prototype.helloStream = function helloStream(requestMessage, metadata) {
  var listeners = {
    data: [],
    end: [],
    status: []
  };
  var client = grpc.invoke(Api.HelloStream, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onMessage: function (responseMessage) {
      listeners.data.forEach(function (handler) {
        handler(responseMessage);
      });
    },
    onEnd: function (status, statusMessage, trailers) {
      listeners.status.forEach(function (handler) {
        handler({ code: status, details: statusMessage, metadata: trailers });
      });
      listeners.end.forEach(function (handler) {
        handler({ code: status, details: statusMessage, metadata: trailers });
      });
      listeners = null;
    }
  });
  return {
    on: function (type, handler) {
      listeners[type].push(handler);
      return this;
    },
    cancel: function () {
      listeners = null;
      client.close();
    }
  };
};

ApiClient.prototype.stdinStream = function stdinStream(requestMessage, metadata) {
  var listeners = {
    data: [],
    end: [],
    status: []
  };
  var client = grpc.invoke(Api.StdinStream, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onMessage: function (responseMessage) {
      listeners.data.forEach(function (handler) {
        handler(responseMessage);
      });
    },
    onEnd: function (status, statusMessage, trailers) {
      listeners.status.forEach(function (handler) {
        handler({ code: status, details: statusMessage, metadata: trailers });
      });
      listeners.end.forEach(function (handler) {
        handler({ code: status, details: statusMessage, metadata: trailers });
      });
      listeners = null;
    }
  });
  return {
    on: function (type, handler) {
      listeners[type].push(handler);
      return this;
    },
    cancel: function () {
      listeners = null;
      client.close();
    }
  };
};

ApiClient.prototype.htopStream = function htopStream(requestMessage, metadata) {
  var listeners = {
    data: [],
    end: [],
    status: []
  };
  var client = grpc.invoke(Api.HtopStream, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onMessage: function (responseMessage) {
      listeners.data.forEach(function (handler) {
        handler(responseMessage);
      });
    },
    onEnd: function (status, statusMessage, trailers) {
      listeners.status.forEach(function (handler) {
        handler({ code: status, details: statusMessage, metadata: trailers });
      });
      listeners.end.forEach(function (handler) {
        handler({ code: status, details: statusMessage, metadata: trailers });
      });
      listeners = null;
    }
  });
  return {
    on: function (type, handler) {
      listeners[type].push(handler);
      return this;
    },
    cancel: function () {
      listeners = null;
      client.close();
    }
  };
};

exports.ApiClient = ApiClient;

var BidiStream = (function () {
  function BidiStream() {}
  BidiStream.serviceName = "BidiStream";
  return BidiStream;
}());

BidiStream.Send = {
  methodName: "Send",
  service: BidiStream,
  requestStream: true,
  responseStream: true,
  requestType: api_pb.Message,
  responseType: api_pb.Message
};

exports.BidiStream = BidiStream;

function BidiStreamClient(serviceHost, options) {
  this.serviceHost = serviceHost;
  this.options = options || {};
}

BidiStreamClient.prototype.send = function send(metadata) {
  var listeners = {
    data: [],
    end: [],
    status: []
  };
  var client = grpc.client(BidiStream.Send, {
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport
  });
  client.onEnd(function (status, statusMessage, trailers) {
    listeners.status.forEach(function (handler) {
      handler({ code: status, details: statusMessage, metadata: trailers });
    });
    listeners.end.forEach(function (handler) {
      handler({ code: status, details: statusMessage, metadata: trailers });
    });
    listeners = null;
  });
  client.onMessage(function (message) {
    listeners.data.forEach(function (handler) {
      handler(message);
    })
  });
  client.start(metadata);
  return {
    on: function (type, handler) {
      listeners[type].push(handler);
      return this;
    },
    write: function (requestMessage) {
      client.send(requestMessage);
      return this;
    },
    end: function () {
      client.finishSend();
    },
    cancel: function () {
      listeners = null;
      client.close();
    }
  };
};

exports.BidiStreamClient = BidiStreamClient;


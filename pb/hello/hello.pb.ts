/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../fetch.pb"
import * as GoogleProtobufEmpty from "../google/protobuf/empty.pb"
export type HelloRequest = {
  name?: string
}

export type HelloResponse = {
  success?: boolean
  message?: string
  name?: string
}

export type UpdateStatusRequest = {
  offerId?: string
  statusId?: string
  note?: string
}

export type UpdateStatusResponse = {
  offerId?: string
}

export class Greeter {
  static Hello(req: HelloRequest, initReq?: fm.InitReq): Promise<HelloResponse> {
    return fm.fetchReq<HelloRequest, HelloResponse>(`/v1/message`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static Ping(req: GoogleProtobufEmpty.Empty, initReq?: fm.InitReq): Promise<HelloResponse> {
    return fm.fetchReq<GoogleProtobufEmpty.Empty, HelloResponse>(`/v1/ping?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static Ping2(req: GoogleProtobufEmpty.Empty, initReq?: fm.InitReq): Promise<HelloResponse> {
    return fm.fetchReq<GoogleProtobufEmpty.Empty, HelloResponse>(`/hello.Greeter/Ping2`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
}
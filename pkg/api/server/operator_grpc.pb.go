//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.1
// source: pkg/api/server/operator.proto

package server

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// OperatorClient is the client API for Operator service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type OperatorClient interface {
	GetVersion(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Version, error)
}

type operatorClient struct {
	cc grpc.ClientConnInterface
}

func NewOperatorClient(cc grpc.ClientConnInterface) OperatorClient {
	return &operatorClient{cc}
}

func (c *operatorClient) GetVersion(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Version, error) {
	out := new(Version)
	err := c.cc.Invoke(ctx, "/server.Operator/GetVersion", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OperatorServer is the server API for Operator service.
// All implementations must embed UnimplementedOperatorServer
// for forward compatibility
type OperatorServer interface {
	GetVersion(context.Context, *Empty) (*Version, error)
	mustEmbedUnimplementedOperatorServer()
}

// UnimplementedOperatorServer must be embedded to have forward compatible implementations.
type UnimplementedOperatorServer struct {
}

func (UnimplementedOperatorServer) GetVersion(context.Context, *Empty) (*Version, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetVersion not implemented")
}
func (UnimplementedOperatorServer) mustEmbedUnimplementedOperatorServer() {}

// UnsafeOperatorServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to OperatorServer will
// result in compilation errors.
type UnsafeOperatorServer interface {
	mustEmbedUnimplementedOperatorServer()
}

func RegisterOperatorServer(s grpc.ServiceRegistrar, srv OperatorServer) {
	s.RegisterService(&Operator_ServiceDesc, srv)
}

func _Operator_GetVersion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OperatorServer).GetVersion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/server.Operator/GetVersion",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OperatorServer).GetVersion(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// Operator_ServiceDesc is the grpc.ServiceDesc for Operator service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Operator_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "server.Operator",
	HandlerType: (*OperatorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetVersion",
			Handler:    _Operator_GetVersion_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/api/server/operator.proto",
}

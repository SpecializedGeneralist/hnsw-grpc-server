// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package grpcapi

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// ServerClient is the client API for Server service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ServerClient interface {
	// CreateIndex makes a new index.
	CreateIndex(ctx context.Context, in *CreateIndexRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// DeleteIndex removes an index.
	DeleteIndex(ctx context.Context, in *DeleteIndexRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// InsertVector inserts a new vector in the given index.
	InsertVector(ctx context.Context, in *InsertVectorRequest, opts ...grpc.CallOption) (*InsertVectorReply, error)
	// InsertVectors inserts the new vectors in the given index. It flushes the index at each batch.
	InsertVectors(ctx context.Context, opts ...grpc.CallOption) (Server_InsertVectorsClient, error)
	// InsertVectorWithID inserts a new vector in the given index.
	InsertVectorWithId(ctx context.Context, in *InsertVectorWithIdRequest, opts ...grpc.CallOption) (*InsertVectorWithIdReply, error)
	// InsertVectorsWithIDs inserts the new vectors in the given index. It flushes the index at each batch.
	InsertVectorsWithIds(ctx context.Context, opts ...grpc.CallOption) (Server_InsertVectorsWithIdsClient, error)
	// SearchKNN returns the top k nearest neighbors to the query, searching on the given index.
	SearchKNN(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (*SearchKNNReply, error)
	// FlushIndex the index to file.
	FlushIndex(ctx context.Context, in *FlushRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// Indices returns the list of indices.
	Indices(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*IndicesReply, error)
	// SetEf sets the `ef` parameter for the given index.
	SetEf(ctx context.Context, in *SetEfRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type serverClient struct {
	cc grpc.ClientConnInterface
}

func NewServerClient(cc grpc.ClientConnInterface) ServerClient {
	return &serverClient{cc}
}

func (c *serverClient) CreateIndex(ctx context.Context, in *CreateIndexRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/grpcapi.Server/CreateIndex", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serverClient) DeleteIndex(ctx context.Context, in *DeleteIndexRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/grpcapi.Server/DeleteIndex", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serverClient) InsertVector(ctx context.Context, in *InsertVectorRequest, opts ...grpc.CallOption) (*InsertVectorReply, error) {
	out := new(InsertVectorReply)
	err := c.cc.Invoke(ctx, "/grpcapi.Server/InsertVector", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serverClient) InsertVectors(ctx context.Context, opts ...grpc.CallOption) (Server_InsertVectorsClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Server_serviceDesc.Streams[0], "/grpcapi.Server/InsertVectors", opts...)
	if err != nil {
		return nil, err
	}
	x := &serverInsertVectorsClient{stream}
	return x, nil
}

type Server_InsertVectorsClient interface {
	Send(*InsertVectorRequest) error
	CloseAndRecv() (*InsertVectorsReply, error)
	grpc.ClientStream
}

type serverInsertVectorsClient struct {
	grpc.ClientStream
}

func (x *serverInsertVectorsClient) Send(m *InsertVectorRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *serverInsertVectorsClient) CloseAndRecv() (*InsertVectorsReply, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(InsertVectorsReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *serverClient) InsertVectorWithId(ctx context.Context, in *InsertVectorWithIdRequest, opts ...grpc.CallOption) (*InsertVectorWithIdReply, error) {
	out := new(InsertVectorWithIdReply)
	err := c.cc.Invoke(ctx, "/grpcapi.Server/InsertVectorWithId", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serverClient) InsertVectorsWithIds(ctx context.Context, opts ...grpc.CallOption) (Server_InsertVectorsWithIdsClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Server_serviceDesc.Streams[1], "/grpcapi.Server/InsertVectorsWithIds", opts...)
	if err != nil {
		return nil, err
	}
	x := &serverInsertVectorsWithIdsClient{stream}
	return x, nil
}

type Server_InsertVectorsWithIdsClient interface {
	Send(*InsertVectorWithIdRequest) error
	CloseAndRecv() (*InsertVectorsWithIdsReply, error)
	grpc.ClientStream
}

type serverInsertVectorsWithIdsClient struct {
	grpc.ClientStream
}

func (x *serverInsertVectorsWithIdsClient) Send(m *InsertVectorWithIdRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *serverInsertVectorsWithIdsClient) CloseAndRecv() (*InsertVectorsWithIdsReply, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(InsertVectorsWithIdsReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *serverClient) SearchKNN(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (*SearchKNNReply, error) {
	out := new(SearchKNNReply)
	err := c.cc.Invoke(ctx, "/grpcapi.Server/SearchKNN", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serverClient) FlushIndex(ctx context.Context, in *FlushRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/grpcapi.Server/FlushIndex", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serverClient) Indices(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*IndicesReply, error) {
	out := new(IndicesReply)
	err := c.cc.Invoke(ctx, "/grpcapi.Server/Indices", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serverClient) SetEf(ctx context.Context, in *SetEfRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/grpcapi.Server/SetEf", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ServerServer is the server API for Server service.
// All implementations must embed UnimplementedServerServer
// for forward compatibility
type ServerServer interface {
	// CreateIndex makes a new index.
	CreateIndex(context.Context, *CreateIndexRequest) (*emptypb.Empty, error)
	// DeleteIndex removes an index.
	DeleteIndex(context.Context, *DeleteIndexRequest) (*emptypb.Empty, error)
	// InsertVector inserts a new vector in the given index.
	InsertVector(context.Context, *InsertVectorRequest) (*InsertVectorReply, error)
	// InsertVectors inserts the new vectors in the given index. It flushes the index at each batch.
	InsertVectors(Server_InsertVectorsServer) error
	// InsertVectorWithID inserts a new vector in the given index.
	InsertVectorWithId(context.Context, *InsertVectorWithIdRequest) (*InsertVectorWithIdReply, error)
	// InsertVectorsWithIDs inserts the new vectors in the given index. It flushes the index at each batch.
	InsertVectorsWithIds(Server_InsertVectorsWithIdsServer) error
	// SearchKNN returns the top k nearest neighbors to the query, searching on the given index.
	SearchKNN(context.Context, *SearchRequest) (*SearchKNNReply, error)
	// FlushIndex the index to file.
	FlushIndex(context.Context, *FlushRequest) (*emptypb.Empty, error)
	// Indices returns the list of indices.
	Indices(context.Context, *emptypb.Empty) (*IndicesReply, error)
	// SetEf sets the `ef` parameter for the given index.
	SetEf(context.Context, *SetEfRequest) (*emptypb.Empty, error)
	mustEmbedUnimplementedServerServer()
}

// UnimplementedServerServer must be embedded to have forward compatible implementations.
type UnimplementedServerServer struct {
}

func (UnimplementedServerServer) CreateIndex(context.Context, *CreateIndexRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateIndex not implemented")
}
func (UnimplementedServerServer) DeleteIndex(context.Context, *DeleteIndexRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteIndex not implemented")
}
func (UnimplementedServerServer) InsertVector(context.Context, *InsertVectorRequest) (*InsertVectorReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InsertVector not implemented")
}
func (UnimplementedServerServer) InsertVectors(Server_InsertVectorsServer) error {
	return status.Errorf(codes.Unimplemented, "method InsertVectors not implemented")
}
func (UnimplementedServerServer) InsertVectorWithId(context.Context, *InsertVectorWithIdRequest) (*InsertVectorWithIdReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InsertVectorWithId not implemented")
}
func (UnimplementedServerServer) InsertVectorsWithIds(Server_InsertVectorsWithIdsServer) error {
	return status.Errorf(codes.Unimplemented, "method InsertVectorsWithIds not implemented")
}
func (UnimplementedServerServer) SearchKNN(context.Context, *SearchRequest) (*SearchKNNReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SearchKNN not implemented")
}
func (UnimplementedServerServer) FlushIndex(context.Context, *FlushRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FlushIndex not implemented")
}
func (UnimplementedServerServer) Indices(context.Context, *emptypb.Empty) (*IndicesReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Indices not implemented")
}
func (UnimplementedServerServer) SetEf(context.Context, *SetEfRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetEf not implemented")
}
func (UnimplementedServerServer) mustEmbedUnimplementedServerServer() {}

// UnsafeServerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ServerServer will
// result in compilation errors.
type UnsafeServerServer interface {
	mustEmbedUnimplementedServerServer()
}

func RegisterServerServer(s grpc.ServiceRegistrar, srv ServerServer) {
	s.RegisterService(&_Server_serviceDesc, srv)
}

func _Server_CreateIndex_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateIndexRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServerServer).CreateIndex(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpcapi.Server/CreateIndex",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServerServer).CreateIndex(ctx, req.(*CreateIndexRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Server_DeleteIndex_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteIndexRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServerServer).DeleteIndex(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpcapi.Server/DeleteIndex",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServerServer).DeleteIndex(ctx, req.(*DeleteIndexRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Server_InsertVector_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InsertVectorRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServerServer).InsertVector(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpcapi.Server/InsertVector",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServerServer).InsertVector(ctx, req.(*InsertVectorRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Server_InsertVectors_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ServerServer).InsertVectors(&serverInsertVectorsServer{stream})
}

type Server_InsertVectorsServer interface {
	SendAndClose(*InsertVectorsReply) error
	Recv() (*InsertVectorRequest, error)
	grpc.ServerStream
}

type serverInsertVectorsServer struct {
	grpc.ServerStream
}

func (x *serverInsertVectorsServer) SendAndClose(m *InsertVectorsReply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *serverInsertVectorsServer) Recv() (*InsertVectorRequest, error) {
	m := new(InsertVectorRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Server_InsertVectorWithId_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InsertVectorWithIdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServerServer).InsertVectorWithId(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpcapi.Server/InsertVectorWithId",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServerServer).InsertVectorWithId(ctx, req.(*InsertVectorWithIdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Server_InsertVectorsWithIds_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ServerServer).InsertVectorsWithIds(&serverInsertVectorsWithIdsServer{stream})
}

type Server_InsertVectorsWithIdsServer interface {
	SendAndClose(*InsertVectorsWithIdsReply) error
	Recv() (*InsertVectorWithIdRequest, error)
	grpc.ServerStream
}

type serverInsertVectorsWithIdsServer struct {
	grpc.ServerStream
}

func (x *serverInsertVectorsWithIdsServer) SendAndClose(m *InsertVectorsWithIdsReply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *serverInsertVectorsWithIdsServer) Recv() (*InsertVectorWithIdRequest, error) {
	m := new(InsertVectorWithIdRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Server_SearchKNN_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServerServer).SearchKNN(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpcapi.Server/SearchKNN",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServerServer).SearchKNN(ctx, req.(*SearchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Server_FlushIndex_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FlushRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServerServer).FlushIndex(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpcapi.Server/FlushIndex",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServerServer).FlushIndex(ctx, req.(*FlushRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Server_Indices_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServerServer).Indices(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpcapi.Server/Indices",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServerServer).Indices(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Server_SetEf_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetEfRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServerServer).SetEf(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpcapi.Server/SetEf",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServerServer).SetEf(ctx, req.(*SetEfRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Server_serviceDesc = grpc.ServiceDesc{
	ServiceName: "grpcapi.Server",
	HandlerType: (*ServerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateIndex",
			Handler:    _Server_CreateIndex_Handler,
		},
		{
			MethodName: "DeleteIndex",
			Handler:    _Server_DeleteIndex_Handler,
		},
		{
			MethodName: "InsertVector",
			Handler:    _Server_InsertVector_Handler,
		},
		{
			MethodName: "InsertVectorWithId",
			Handler:    _Server_InsertVectorWithId_Handler,
		},
		{
			MethodName: "SearchKNN",
			Handler:    _Server_SearchKNN_Handler,
		},
		{
			MethodName: "FlushIndex",
			Handler:    _Server_FlushIndex_Handler,
		},
		{
			MethodName: "Indices",
			Handler:    _Server_Indices_Handler,
		},
		{
			MethodName: "SetEf",
			Handler:    _Server_SetEf_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "InsertVectors",
			Handler:       _Server_InsertVectors_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "InsertVectorsWithIds",
			Handler:       _Server_InsertVectorsWithIds_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "hnswservice.proto",
}

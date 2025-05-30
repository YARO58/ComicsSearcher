// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.2
// source: proto/words/words.proto

package words

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	Words_Ping_FullMethodName = "/words.Words/Ping"
	Words_Norm_FullMethodName = "/words.Words/Norm"
)

// WordsClient is the client API for Words service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type WordsClient interface {
	Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
	Norm(ctx context.Context, in *WordsRequest, opts ...grpc.CallOption) (*WordsReply, error)
}

type wordsClient struct {
	cc grpc.ClientConnInterface
}

func NewWordsClient(cc grpc.ClientConnInterface) WordsClient {
	return &wordsClient{cc}
}

func (c *wordsClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Words_Ping_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *wordsClient) Norm(ctx context.Context, in *WordsRequest, opts ...grpc.CallOption) (*WordsReply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(WordsReply)
	err := c.cc.Invoke(ctx, Words_Norm_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WordsServer is the server API for Words service.
// All implementations must embed UnimplementedWordsServer
// for forward compatibility.
type WordsServer interface {
	Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	Norm(context.Context, *WordsRequest) (*WordsReply, error)
	mustEmbedUnimplementedWordsServer()
}

// UnimplementedWordsServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedWordsServer struct{}

func (UnimplementedWordsServer) Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedWordsServer) Norm(context.Context, *WordsRequest) (*WordsReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Norm not implemented")
}
func (UnimplementedWordsServer) mustEmbedUnimplementedWordsServer() {}
func (UnimplementedWordsServer) testEmbeddedByValue()               {}

// UnsafeWordsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WordsServer will
// result in compilation errors.
type UnsafeWordsServer interface {
	mustEmbedUnimplementedWordsServer()
}

func RegisterWordsServer(s grpc.ServiceRegistrar, srv WordsServer) {
	// If the following call pancis, it indicates UnimplementedWordsServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Words_ServiceDesc, srv)
}

func _Words_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WordsServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Words_Ping_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WordsServer).Ping(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Words_Norm_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WordsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WordsServer).Norm(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Words_Norm_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WordsServer).Norm(ctx, req.(*WordsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Words_ServiceDesc is the grpc.ServiceDesc for Words service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Words_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "words.Words",
	HandlerType: (*WordsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _Words_Ping_Handler,
		},
		{
			MethodName: "Norm",
			Handler:    _Words_Norm_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/words/words.proto",
}

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: svc/compute/billingaccount/billingaccount.proto

package billingaccount

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

// BillingAccountServiceClient is the client API for BillingAccountService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BillingAccountServiceClient interface {
	CreateBillingAccount(ctx context.Context, in *CreateBillingAccountRequest, opts ...grpc.CallOption) (*BillingAccount, error)
	GetBillingAccount(ctx context.Context, in *GetBillingAccountRequest, opts ...grpc.CallOption) (*BillingAccount, error)
	ListBillingAccounts(ctx context.Context, in *ListBillingAccountsRequest, opts ...grpc.CallOption) (*ListBillingAccountsResponse, error)
}

type billingAccountServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewBillingAccountServiceClient(cc grpc.ClientConnInterface) BillingAccountServiceClient {
	return &billingAccountServiceClient{cc}
}

func (c *billingAccountServiceClient) CreateBillingAccount(ctx context.Context, in *CreateBillingAccountRequest, opts ...grpc.CallOption) (*BillingAccount, error) {
	out := new(BillingAccount)
	err := c.cc.Invoke(ctx, "/org.cudo.compute.v1.BillingAccountService/CreateBillingAccount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *billingAccountServiceClient) GetBillingAccount(ctx context.Context, in *GetBillingAccountRequest, opts ...grpc.CallOption) (*BillingAccount, error) {
	out := new(BillingAccount)
	err := c.cc.Invoke(ctx, "/org.cudo.compute.v1.BillingAccountService/GetBillingAccount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *billingAccountServiceClient) ListBillingAccounts(ctx context.Context, in *ListBillingAccountsRequest, opts ...grpc.CallOption) (*ListBillingAccountsResponse, error) {
	out := new(ListBillingAccountsResponse)
	err := c.cc.Invoke(ctx, "/org.cudo.compute.v1.BillingAccountService/ListBillingAccounts", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BillingAccountServiceServer is the server API for BillingAccountService service.
// All implementations must embed UnimplementedBillingAccountServiceServer
// for forward compatibility
type BillingAccountServiceServer interface {
	CreateBillingAccount(context.Context, *CreateBillingAccountRequest) (*BillingAccount, error)
	GetBillingAccount(context.Context, *GetBillingAccountRequest) (*BillingAccount, error)
	ListBillingAccounts(context.Context, *ListBillingAccountsRequest) (*ListBillingAccountsResponse, error)
	mustEmbedUnimplementedBillingAccountServiceServer()
}

// UnimplementedBillingAccountServiceServer must be embedded to have forward compatible implementations.
type UnimplementedBillingAccountServiceServer struct {
}

func (UnimplementedBillingAccountServiceServer) CreateBillingAccount(context.Context, *CreateBillingAccountRequest) (*BillingAccount, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateBillingAccount not implemented")
}
func (UnimplementedBillingAccountServiceServer) GetBillingAccount(context.Context, *GetBillingAccountRequest) (*BillingAccount, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBillingAccount not implemented")
}
func (UnimplementedBillingAccountServiceServer) ListBillingAccounts(context.Context, *ListBillingAccountsRequest) (*ListBillingAccountsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListBillingAccounts not implemented")
}
func (UnimplementedBillingAccountServiceServer) mustEmbedUnimplementedBillingAccountServiceServer() {}

// UnsafeBillingAccountServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BillingAccountServiceServer will
// result in compilation errors.
type UnsafeBillingAccountServiceServer interface {
	mustEmbedUnimplementedBillingAccountServiceServer()
}

func RegisterBillingAccountServiceServer(s grpc.ServiceRegistrar, srv BillingAccountServiceServer) {
	s.RegisterService(&BillingAccountService_ServiceDesc, srv)
}

func _BillingAccountService_CreateBillingAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateBillingAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BillingAccountServiceServer).CreateBillingAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/org.cudo.compute.v1.BillingAccountService/CreateBillingAccount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BillingAccountServiceServer).CreateBillingAccount(ctx, req.(*CreateBillingAccountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BillingAccountService_GetBillingAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBillingAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BillingAccountServiceServer).GetBillingAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/org.cudo.compute.v1.BillingAccountService/GetBillingAccount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BillingAccountServiceServer).GetBillingAccount(ctx, req.(*GetBillingAccountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BillingAccountService_ListBillingAccounts_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListBillingAccountsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BillingAccountServiceServer).ListBillingAccounts(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/org.cudo.compute.v1.BillingAccountService/ListBillingAccounts",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BillingAccountServiceServer).ListBillingAccounts(ctx, req.(*ListBillingAccountsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// BillingAccountService_ServiceDesc is the grpc.ServiceDesc for BillingAccountService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BillingAccountService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "org.cudo.compute.v1.BillingAccountService",
	HandlerType: (*BillingAccountServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateBillingAccount",
			Handler:    _BillingAccountService_CreateBillingAccount_Handler,
		},
		{
			MethodName: "GetBillingAccount",
			Handler:    _BillingAccountService_GetBillingAccount_Handler,
		},
		{
			MethodName: "ListBillingAccounts",
			Handler:    _BillingAccountService_ListBillingAccounts_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "svc/compute/billingaccount/billingaccount.proto",
}

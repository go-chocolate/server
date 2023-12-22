package interceptor

//// TraceUnaryClientInterceptor
//func TraceUnaryClientInterceptor() grpc.UnaryClientInterceptor {
//	interceptor := otelgrpc.UnaryClientInterceptor()
//	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
//		return interceptor(ctx, method, req, reply, cc, invoker, opts...)
//	}
//}
//
//// TraceStreamClientInterceptor
//func TraceStreamClientInterceptor() grpc.StreamClientInterceptor {
//	interceptor := otelgrpc.StreamClientInterceptor()
//	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
//		return interceptor(ctx, desc, cc, method, streamer, opts...)
//	}
//}

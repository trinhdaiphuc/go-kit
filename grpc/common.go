package grpc

import "strings"

func SplitGRPCMethodName(fullMethodName string) (serviceName, methodName string) {
	fullMethodName = strings.TrimPrefix(fullMethodName, "/") // remove leading slash
	if before, after, ok := strings.Cut(fullMethodName, "/"); ok {
		return before, after
	}
	return "unknown", "unknown"
}

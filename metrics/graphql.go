package metrics

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

type Query struct {
	OperationName string `json:"operationName"`
	Query         string `json:"Query"`
}

func GraphqlInterceptor(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	rc := graphql.GetOperationContext(ctx)
	startTime := rc.Stats.OperationStart

	resp := next(ctx)

	latency := time.Since(startTime)
	method := ""
	endpoint := ""
	status := "200"
	if rc.Operation != nil {
		method = string(rc.Operation.Operation)
		endpoint = rc.Operation.Name
	}

	if len(resp.Errors) > 0 {
		status = "500"
	}

	doneHTTPHandleRequest(InboundCall, method, endpoint, status, latency.Seconds())
	return resp
}

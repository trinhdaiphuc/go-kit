package metrics

import (
	"context"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const (
	availableLabelMethod = "available"
	consumeLabelMethod   = "consume"
)

type Session interface {
	Run(ctx context.Context, cypher string, params map[string]any, configurers ...func(*neo4j.TransactionConfig)) (neo4j.ResultWithContext, error)
}

type RunFunc = func(ctx context.Context, cypher string, params map[string]any, configurers ...func(*neo4j.TransactionConfig)) (neo4j.ResultWithContext, error)

func Neo4jRunInterceptor(session Session, queryName string) RunFunc {
	return func(ctx context.Context, cypher string, params map[string]any, configurers ...func(*neo4j.TransactionConfig)) (neo4j.ResultWithContext, error) {
		startTime := time.Now()
		run, err := session.Run(ctx, cypher, params, configurers...)
		elapsedTime := time.Since(startTime).Seconds()

		statusCode := "200" // Success
		if err != nil {
			statusCode = "500" // Error
		}

		doneHandleRequest(ClientCall, databaseLabelMethod, queryName, statusCode, statusCode, elapsedTime)
		return run, err
	}
}

func Neo4jManagedTransactionWork(queryName string, f func(tx neo4j.ManagedTransaction) (any, error)) neo4j.ManagedTransactionWork {
	return func(tx neo4j.ManagedTransaction) (any, error) {
		startTime := time.Now()
		run, err := f(tx)
		elapsedTime := time.Since(startTime).Seconds()

		statusCode := "200" // Success
		if err != nil {
			statusCode = "500" // Error
		}

		doneHandleRequest(ClientCall, databaseLabelMethod, queryName, statusCode, statusCode, elapsedTime)
		return run, err
	}
}

func ObserveNeo4jExecution(queryName string, summary neo4j.ResultSummary, err error) {
	statusCode := "200"

	if err != nil {
		statusCode = "500"
	}

	doneHandleRequest(ClientCall, availableLabelMethod, queryName, statusCode, statusCode, summary.ResultAvailableAfter().Seconds())
	doneHandleRequest(ClientCall, consumeLabelMethod, queryName, statusCode, statusCode, summary.ResultConsumedAfter().Seconds())
}

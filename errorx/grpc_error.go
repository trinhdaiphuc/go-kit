package errorx

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ConvertGRPCError(err error) error {
	if err == nil {
		return nil
	}

	errWrapper := &ErrorWrapper{}

	ok := errors.As(err, &errWrapper)
	if ok {
		if errWrapper.GetStatus() == codes.OK {
			return nil
		}
		return status.Error(errWrapper.GetStatus(), errWrapper.GetMessage())
	}

	return status.Error(codes.Unknown, err.Error())
}

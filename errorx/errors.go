package errorx

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/trinhdaiphuc/go-kit/log"
)

type IsErrorDetail interface {
	isErrorDetail()
}

type ErrorWrapper struct {
	cause     error     `json:"-"` // The original error that caused this error to be returned.
	ErrorBody ErrorBody `json:"error"`
}

// ErrorBody is the structure of the error response. Follow https://google.aip.dev/193#http11json-representation
type ErrorBody struct {
	Code       int          `json:"code"`              // The HTTP status code that corresponds to `google.rpc.Status.code`.
	Status     string       `json:"status,omitempty"`  // This corresponds to `google.rpc.Status.code`.
	Message    string       `json:"message"`           // This corresponds to `google.rpc.Status.message`.
	Details    *ErrorDetail `json:"details,omitempty"` // This corresponds to `google.rpc.Status.details`.
	StatusCode codes.Code   `json:"-"`                 // The HTTP status code that corresponds to `google.rpc.Status.code`.
}

type ErrorDetail struct {
	ErrorInfo           *ErrorInfo           `json:"error_info,omitempty"`
	LocalizedMessage    *LocalizedMessage    `json:"localized_message,omitempty"`
	BadRequest          *BadRequest          `json:"bad_request,omitempty"`
	PreconditionFailure *PreconditionFailure `json:"precondition_failure,omitempty"`
	ResourceInfo        *ResourceInfo        `json:"resource_info,omitempty"`
	QuotaFailure        *QuotaFailure        `json:"quota_failure,omitempty"`
	DebugInfo           *DebugInfo           `json:"debug_info,omitempty"`
	Help                *Help                `json:"help,omitempty"`
}

func New(message string) *ErrorWrapper {
	return newError(message)
}

func Newf(format string, args ...any) *ErrorWrapper {
	return newError(fmt.Sprintf(format, args...))
}

func Wrap(err error, message string) *ErrorWrapper {
	if err == nil {
		return newError(message)
	}

	e := newError(message)
	e.cause = err

	return e
}

func Wrapf(err error, format string, args ...any) *ErrorWrapper {
	if err == nil {
		return newError(fmt.Sprintf(format, args...))
	}

	e := newError(fmt.Sprintf(format, args...))
	e.cause = err
	return e
}

// Unwrap returns the original error that caused this error to be returned. Compatible with Go 1.13 error wrapping.
func (e *ErrorWrapper) Unwrap() error {
	if e == nil {
		return nil
	}

	if e.cause == nil {
		return nil
	}

	return e.cause
}

func (e *ErrorWrapper) Error() string {
	if e == nil {
		return "nil error"
	}

	if e.cause == nil {
		return e.ErrorBody.Message
	}

	if e.ErrorBody.Message == "" {
		return e.cause.Error()
	}

	return e.ErrorBody.Message + ": " + e.cause.Error()
}

func (e *ErrorWrapper) WithCode(code int) *ErrorWrapper {
	e.ErrorBody.Code = code
	return e
}

func (e *ErrorWrapper) WithStatus(status codes.Code) *ErrorWrapper {
	e.ErrorBody.StatusCode = status
	e.ErrorBody.Status = CodeToString(status)
	return e
}

func (e *ErrorWrapper) WithMessage(message string) *ErrorWrapper {
	e.ErrorBody.Message = message
	return e
}

func (e *ErrorWrapper) WithCodeFromStatus(status codes.Code) *ErrorWrapper {
	return e.WithStatus(status).WithCode(runtime.HTTPStatusFromCode(status))
}

func (e *ErrorWrapper) WithDetails(details ...IsErrorDetail) *ErrorWrapper {
	if len(details) == 0 {
		return e
	}

	errDetail, err := transformErrDetails(details...)
	if err != nil {
		log.For(nil).Error("transform error details failed", zap.Error(err))
		return e
	}

	e.ErrorBody.Details = errDetail
	return e
}

func (e *ErrorWrapper) GetCode() int {
	if e == nil {
		return http.StatusInternalServerError
	}

	return e.ErrorBody.Code
}

func (e *ErrorWrapper) GetStatus() codes.Code {
	if e == nil {
		return codes.Internal
	}

	return e.ErrorBody.StatusCode
}

func (e *ErrorWrapper) GetMessage() string {
	if e == nil {
		return "nil error"
	}

	if e.ErrorBody.Message == "" {
		return e.Error()
	}

	return e.ErrorBody.Message
}

var codeToString = map[codes.Code]string{
	codes.OK:                 "OK",
	codes.Canceled:           "CANCELLED",
	codes.Unknown:            "UNKNOWN",
	codes.InvalidArgument:    "INVALID_ARGUMENT",
	codes.DeadlineExceeded:   "DEADLINE_EXCEEDED",
	codes.NotFound:           "NOT_FOUND",
	codes.AlreadyExists:      "ALREADY_EXISTS",
	codes.PermissionDenied:   "PERMISSION_DENIED",
	codes.ResourceExhausted:  "RESOURCE_EXHAUSTED",
	codes.FailedPrecondition: "FAILED_PRECONDITION",
	codes.Aborted:            "ABORTED",
	codes.OutOfRange:         "OUT_OF_RANGE",
	codes.Unimplemented:      "UNIMPLEMENTED",
	codes.Internal:           "INTERNAL",
	codes.Unavailable:        "UNAVAILABLE",
	codes.DataLoss:           "DATA_LOSS",
	codes.Unauthenticated:    "UNAUTHENTICATED",
}

func CodeToString(c codes.Code) string {
	str, ok := codeToString[c]
	if ok {
		return str
	}

	return "CODE(" + strconv.FormatInt(int64(c), 10) + ")"
}

// newError is a function to create a new ErrorWrapper default error code
func newError(message string) *ErrorWrapper {
	return &ErrorWrapper{
		ErrorBody: ErrorBody{
			Code:       http.StatusInternalServerError,
			StatusCode: codes.Internal,
			Message:    message,
		},
	}
}

func ParseValidateDetails(err error) *BadRequest {
	var validateErr validator.ValidationErrors
	ok := errors.As(err, &validateErr)
	if !ok || validateErr == nil {
		return nil
	}

	result := &BadRequest{}
	for _, e := range validateErr {
		description := fmt.Sprintf("Key: '%s' failed on the '%s' tag.", e.Field(), e.Tag())
		if len(e.Param()) > 0 {
			description += " Accepted values: [" + strings.Join(strings.Split(e.Param(), " "), ", ") + "]"
		}
		result.FieldViolations = append(
			result.FieldViolations, &BadRequestFieldViolation{
				Field:       e.Field(),
				Description: description,
			},
		)
	}
	return result
}

type ErrorInfo struct {
	Reason   string         `json:"reason,omitempty"`
	Domain   string         `json:"domain,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

func (e *ErrorInfo) isErrorDetail() {}

type LocalizedMessage struct {
	Locale  string `json:"locale,omitempty"`
	Message string `json:"message,omitempty"`
}

func (l *LocalizedMessage) isErrorDetail() {}

type BadRequest struct {
	FieldViolations []*BadRequestFieldViolation `json:"field_violations,omitempty"`
}

func (b *BadRequest) isErrorDetail() {}

type BadRequestFieldViolation struct {
	Field       string `json:"field,omitempty"`
	Description string `json:"description,omitempty"`
}

type PreconditionFailure struct {
	Violations []*PreconditionFailureViolation `json:"violations,omitempty"`
}

func (p *PreconditionFailure) isErrorDetail() {}

type PreconditionFailureViolation struct {
	Type        string `json:"type,omitempty"`
	Subject     string `json:"subject,omitempty"`
	Description string `json:"description,omitempty"`
}

func (e *PreconditionFailureViolation) isErrorDetail() {}

type ResourceInfo struct {
	ResourceType string `protobuf:"bytes,1,opt,name=resource_type,json=resourceType,proto3" json:"resource_type,omitempty"`
	ResourceName string `protobuf:"bytes,2,opt,name=resource_name,json=resourceName,proto3" json:"resource_name,omitempty"`
	Owner        string `protobuf:"bytes,3,opt,name=owner,proto3" json:"owner,omitempty"`
	Description  string `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
}

func (r *ResourceInfo) isErrorDetail() {}

type QuotaFailure struct {
	Violations []*QuotaFailureViolation `protobuf:"bytes,1,rep,name=violations,proto3" json:"violations,omitempty"`
}

func (q *QuotaFailure) isErrorDetail() {}

type QuotaFailureViolation struct {
	Subject     string `protobuf:"bytes,1,opt,name=subject,proto3" json:"subject,omitempty"`
	Description string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
}

type DebugInfo struct {
	StackEntries []string `protobuf:"bytes,1,rep,name=stack_entries,json=stackEntries,proto3" json:"stack_entries,omitempty"`
	Detail       string   `protobuf:"bytes,2,opt,name=detail,proto3" json:"detail,omitempty"`
}

type Help struct {
	Links []*HelpLink `json:"links,omitempty"`
}

func (h Help) isErrorDetail() {}

type HelpLink struct {
	Description string `json:"description,omitempty"`
	Url         string `json:"url,omitempty"`
}

func (d *DebugInfo) isErrorDetail() {}

var (
	_ IsErrorDetail = (*ErrorInfo)(nil)
	_ IsErrorDetail = (*LocalizedMessage)(nil)
	_ IsErrorDetail = (*BadRequest)(nil)
	_ IsErrorDetail = (*PreconditionFailure)(nil)
	_ IsErrorDetail = (*ResourceInfo)(nil)
	_ IsErrorDetail = (*QuotaFailure)(nil)
	_ IsErrorDetail = (*DebugInfo)(nil)
	_ IsErrorDetail = (*Help)(nil)
)

func GetGRPCCode(err error) codes.Code {
	s, ok := status.FromError(err)
	if ok {
		return s.Code()
	}

	var errWrapper *ErrorWrapper

	ok = errors.As(err, &errWrapper)
	if ok {
		return errWrapper.ErrorBody.StatusCode
	}

	return codes.Unknown
}

func GetHTTPCode(err error) int {
	var (
		errWrapper *ErrorWrapper
		statusCode = http.StatusInternalServerError
	)
	ok := errors.As(err, &errWrapper)
	if ok {
		statusCode = int(errWrapper.ErrorBody.Code)
	}

	return statusCode
}

// AttachContextError attaches an error to the current context. The error is pushed to a list of errors.
func AttachContextError(ctx *gin.Context, logger log.Logger, err error) {
	// This function return error wrapper so that not need to log error again. This log for avoiding lint error
	ctxErr := ctx.Error(err)
	if ctxErr != nil {
		logger.Debug("Attach error into context", zap.Error(err))
	}
}

func convertErrDetails(errDetailItem IsErrorDetail, errDetail *ErrorDetail) {
	switch errDetailItem.(type) {
	case *ErrorInfo:
		errDetail.ErrorInfo = errDetailItem.(*ErrorInfo)
	case *LocalizedMessage:
		errDetail.LocalizedMessage = errDetailItem.(*LocalizedMessage)
	case *BadRequest:
		errDetail.BadRequest = errDetailItem.(*BadRequest)
	case *PreconditionFailure:
		errDetail.PreconditionFailure = errDetailItem.(*PreconditionFailure)
	case *ResourceInfo:
		errDetail.ResourceInfo = errDetailItem.(*ResourceInfo)
	case *QuotaFailure:
		errDetail.QuotaFailure = errDetailItem.(*QuotaFailure)
	case *DebugInfo:
		errDetail.DebugInfo = errDetailItem.(*DebugInfo)
	case *Help:
		errDetail.Help = errDetailItem.(*Help)
	}
}

func transformErrDetails(details ...IsErrorDetail) (*ErrorDetail, error) {
	result := &ErrorDetail{}

	for _, detail := range details {
		convertErrDetails(detail, result)
	}

	return result, nil
}

var (
	ErrorInternal = New("Internal server error")
)

package protocol

import (
	"context"
	"net/http"
	"strconv"

	"github.com/dolab/gogo"
)

type contextKey int

const (
	ctxProtocolKey contextKey = iota + 1
	ctxPackageKey
	ctxServiceKey
	ctxMethodKey
	ctxLoggerKey
	ctxRequestHeaderKey
	ctxStatusCodeKey
	ctxResponseWriterKey
)

func WithPackage(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, ctxPackageKey, name)
}

func ContextPackage(ctx context.Context) (name string, ok bool) {
	name, ok = ctx.Value(ctxPackageKey).(string)
	return
}

func WithService(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, ctxServiceKey, name)
}

func ContextService(ctx context.Context) (name string, ok bool) {
	name, ok = ctx.Value(ctxServiceKey).(string)
	return
}

func WithMethod(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, ctxMethodKey, name)
}

func ContextMethod(ctx context.Context) (name string, ok bool) {
	name, ok = ctx.Value(ctxMethodKey).(string)
	return
}

func WithLogger(ctx context.Context, log gogo.Logger) context.Context {
	return context.WithValue(ctx, ctxLoggerKey, log)
}

func ContextLogger(ctx context.Context) (log gogo.Logger, ok bool) {
	log, ok = ctx.Value(ctxLoggerKey).(gogo.Logger)
	return
}

func WithRequestHeader(ctx context.Context, header http.Header) context.Context {
	return context.WithValue(ctx, ctxRequestHeaderKey, header)
}

func ContextRequestHeader(ctx context.Context) (header http.Header, ok bool) {
	header, ok = ctx.Value(ctxRequestHeaderKey).(http.Header)
	return
}

func WithStatusCode(ctx context.Context, code int) context.Context {
	return context.WithValue(ctx, ctxStatusCodeKey, strconv.Itoa(code))
}

func ContextStatusCode(ctx context.Context) (statusCode int, ok bool) {
	statusCode, ok = ctx.Value(ctxStatusCodeKey).(int)
	return
}

func WithResponseWriter(ctx context.Context, w http.ResponseWriter) context.Context {
	return context.WithValue(ctx, ctxResponseWriterKey, w)
}

func ContextResponseWriter(ctx context.Context) (writer http.ResponseWriter, ok bool) {
	writer, ok = ctx.Value(ctxResponseWriterKey).(http.ResponseWriter)
	return
}

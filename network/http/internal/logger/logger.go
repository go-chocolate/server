package logger

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/go-chocolate/server/netutil"
	"github.com/go-chocolate/server/traceutil"
)

// Logger http日志记录中间件
func Logger(options ...Option) func(next http.Handler) http.Handler {
	var opt = applyLoggerOption(options...)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			var entity = new(Entity)
			entity.Begin = time.Now()
			defer func() {
				if opt.recorder != nil {
					opt.recorder.Record(entity)
				} else {
					logger := logrus.WithContext(entity.Context).WithFields(entity.M())
					if entity.StatusCode >= 200 && entity.StatusCode < 400 {
						logger.Info()
					} else {
						logger.Error()
					}
				}
			}()
			entity.TraceId = traceutil.TraceIDFromContext(request.Context())
			entity.Context = request.Context()
			entity.Host = request.Host
			entity.Url = request.URL
			entity.Path = request.RequestURI
			entity.Method = request.Method
			entity.ClientIp = netutil.ClientIP(request)
			entity.ClientUa = request.UserAgent()
			entity.RequestContentType = request.Header.Get("Content-type")
			entity.RequestReferer = request.Referer()
			entity.RequestContentLength = request.ContentLength
			entity.RequestHeader = map[string]string{}
			for _, key := range opt.requestHeaders {
				entity.RequestHeader[key] = request.Header.Get(key)
			}

			var lw = &logWriter{ResponseWriter: writer}
			writer = lw

			next.ServeHTTP(lw, request)
			entity.Context = request.Context()
			entity.StatusCode = lw.Status()
			entity.ResponseContentType = lw.Header().Get("Content-Type")
			entity.End = time.Now()
			entity.Duration = entity.End.Sub(entity.Begin)
			entity.ResponseContentLength = int64(lw.Length())
			entity.ResponseHeader = map[string]string{}
			rh := lw.Header()
			for _, key := range opt.responseHeaders {
				entity.ResponseHeader[key] = rh.Get(key)
			}
		})
	}
}

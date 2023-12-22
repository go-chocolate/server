package logger

import (
	"context"
	"net/url"
	"time"
)

type Entity struct {
	TraceId               string            // 链路id
	Host                  string            // 请求域名
	Url                   *url.URL          // 请求url
	Path                  string            // 请求路径
	Method                string            // 请求方法
	StatusCode            int               // 返回状态码
	Begin                 time.Time         //
	End                   time.Time         //
	Duration              time.Duration     // 请求耗时
	ClientIp              string            // 客户端ip
	ClientUa              string            // 客户端ua
	RequestReferer        string            // 请求referer
	RequestContentType    string            // 请求类型
	RequestHeader         map[string]string // 请求header
	Request               string            // 格式化后请求参数
	RequestContentLength  int64             // 请求长度
	ResponseContentType   string            // 响应类型
	ResponseHeader        map[string]string // 响应header
	Response              string            // 格式化后响应数据
	ResponseContentLength int64             // 响应长度
	TenantId              string            //
	Context               context.Context   //
}

func (e *Entity) M() map[string]any {
	return map[string]any{
		"traceId":               e.TraceId,
		"host":                  e.Host,
		"url":                   e.Url,
		"path":                  e.Path,
		"method":                e.Method,
		"statusCode":            e.StatusCode,
		"begin":                 e.Begin,
		"end":                   e.End,
		"duration":              e.Duration,
		"clientIp":              e.ClientIp,
		"clientUa":              e.ClientUa,
		"requestReferer":        e.RequestReferer,
		"requestContentType":    e.RequestContentType,
		"requestHeader":         e.RequestHeader,
		"request":               e.Request,
		"requestContentLength":  e.RequestContentLength,
		"responseContentType":   e.ResponseContentType,
		"responseHeader":        e.ResponseHeader,
		"response":              e.Response,
		"responseContentLength": e.ResponseContentLength,
		"tenantId":              e.TenantId,
	}
}

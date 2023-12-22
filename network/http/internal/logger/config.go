package logger

import "strings"

type Option func(o *loggerOption)

type loggerOption struct {
	//recordRequest       bool     //开启请求数据记录
	//recordRequestLimit  int      //请求数据记录长度限制
	//recordResponse      bool     //开启响应数据记录
	//recordResponseLimit int      //响应数据记录长度限制
	requestHeaders  []string //记录请求header
	responseHeaders []string //记录响应header
	ignorePaths     []string //忽略path，命中配置的请求路径不记录
	recorder        Recorder //自定义日志收集器
}

func (o *loggerOption) match(path string) bool {
	for _, v := range o.ignorePaths {
		if v == "" {
			continue
		}
		if v[len(v)-1] == '*' {
			if strings.HasPrefix(path, v[:len(v)-1]) {
				return true
			}
			continue
		}

		if v == path {
			return true
		}
	}
	return false
}

func WithIgnorePath(path ...string) Option {
	return func(o *loggerOption) {
		o.ignorePaths = append(o.ignorePaths, path...)
	}
}

func WithRequestHeader(header ...string) Option {
	return func(o *loggerOption) {
		o.requestHeaders = append(o.requestHeaders, header...)
	}
}

func WithResponseHeader(header ...string) Option {
	return func(o *loggerOption) {
		o.responseHeaders = append(o.responseHeaders, header...)
	}
}

func applyLoggerOption(options ...Option) *loggerOption {
	o := &loggerOption{}
	for _, opt := range options {
		opt(o)
	}
	return o
}

type Recorder interface {
	Record(entity *Entity)
}

type RecordFunc func(entity *Entity)

func (f RecordFunc) Record(entity *Entity) {
	f(entity)
}

var textTypes = []string{
	"application/json",
	"application/xml",
	"application/x-www-form-urlencoded",
	"text/plain",
	"text/xml",
}

func isText(contentType string) bool {
	trueType := strings.TrimSpace(strings.Split(contentType, ";")[0])
	for _, ty := range textTypes {
		if trueType == ty {
			return true
		}
	}
	return false
}

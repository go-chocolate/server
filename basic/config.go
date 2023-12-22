package basic

type Config struct {
	ServiceName  string //服务名
	Addr         string //监听地址加端口
	Tracing      bool   //链路追踪
	Cors         bool   //
	SkipRegistry bool   //
}

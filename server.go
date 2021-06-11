package main

import (
	"github.com/gin-gonic/gin"
	"github.com/wethedevelop/gateway/account"
	"github.com/wethedevelop/gateway/etcdv3"
	"google.golang.org/grpc/resolver"
	"os"
)

// 路由绑定在这里
func Route() *gin.Engine {
	// grpc 负载均衡
	// 创建一个Builder
	builder := etcdv3.NewResolver(os.Getenv("ETCD_ADDR"))
	// 将会以rs.Scheme()为key进行注册
	// 调用 grpc.DialContext(ctx, r.Scheme()+"://authority/"+"serviceName",...)时
	// 会根据第二个参数解析出一个resolver.Target{}
	// Scheme=r.Scheme()
	// Authority="authority"
	// Endpoint="serviceName"
	// 根据Scheme将拿到我们注册的Builder
	resolver.Register(builder)

	r := gin.Default()

	r.POST("/api/user/signup", account.Signup)
	return r
}

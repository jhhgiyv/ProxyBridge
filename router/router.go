package router

import (
	"ProxyBridge/proxy"
	"ProxyBridge/router_struct"
	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	api := r.Group("/api")
	api.Use(authMiddleware())
	proxyGroup := api.Group("proxy")
	proxyGroup.POST("new", NewProxy)
}

func NewProxy(context *gin.Context) {
	resp := make(chan router_struct.Response)
	request := &router_struct.NewProxyRequests{}
	err := context.BindJSON(request)
	if err != nil {
		context.JSON(400, router_struct.Response{Status: 400, Message: "Invalid request\n" + err.Error()})
		return
	}
	go proxy.CreateProxy(request, resp)
	response := <-resp
	context.JSON(response.Status, response)
}

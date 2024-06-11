package main

import (
	"ProxyBridge/config"
	"ProxyBridge/router"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	config.InitConfig()
	r := gin.Default()
	router.InitRouter(r)
	err := r.Run(config.C.GinListen)
	if err != nil {
		log.Fatal(err)
	}
}

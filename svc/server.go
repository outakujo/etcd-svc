package svc

import (
	"etcd-svc/etcd"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func Run(svc string, port int) error {
	engine := gin.New()
	engine.GET("", func(c *gin.Context) {
		c.JSON(http.StatusOK, svc)
	})
	engine.GET("good", func(c *gin.Context) {
		get := etcd.RegisterSvcs.Get("good")
		c.JSON(http.StatusOK, get)
	})
	engine.GET("user", func(c *gin.Context) {
		get := etcd.RegisterSvcs.Get("user")
		c.JSON(http.StatusOK, get)
	})
	return engine.Run(":" + strconv.Itoa(port))
}

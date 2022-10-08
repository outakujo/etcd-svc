package gateway

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func Run(port int) error {
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		match := Router.Match(c.Request)
		fmt.Println(Router)
		c.JSON(http.StatusOK, match)
	})
	return engine.Run(":" + strconv.Itoa(port))
}

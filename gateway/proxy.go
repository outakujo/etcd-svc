package gateway

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func Run(port int) error {
	engine := gin.New()
	engine.Any("/*any", func(c *gin.Context) {
		match := Match(c.Request)
		fmt.Println(Router)
		c.JSON(http.StatusOK, match)
	})
	return engine.Run(":" + strconv.Itoa(port))
}

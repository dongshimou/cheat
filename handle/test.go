package handle

import (
	"cheat/websock"
	"github.com/gin-gonic/gin"
)

func Testws(c *gin.Context) {

	//c.Header("access-control-allow-origin", "*")
	websock.Test(c)

}

func Cheat3(c *gin.Context) {

	websock.Cheat3(c)
}

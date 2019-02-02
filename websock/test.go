package websock

import (
	"cheat/logger"
	"cheat/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Test (c *gin.Context) {

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Warnning("upgrade:", err)
		return
	}
	defer conn.Close()

	service.Test(conn)

}

func Cheat3(c *gin.Context){

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Warnning("upgrade:", err)
		return
	}
	defer conn.Close()

	service.Cheat3(conn)
}
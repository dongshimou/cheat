package main

import (
	"cheat/handle"
	"cheat/logger"
	"github.com/gin-gonic/gin"

)

func init(){
	logger.New("cheat3")
}


func main(){

	router:=gin.Default()

	router.GET("/test",handle.Testws)

	router.GET("/cheat3",handle.Cheat3)

	router.Run(":9090")
}

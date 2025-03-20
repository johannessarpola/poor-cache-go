package rest

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter(rg *gin.RouterGroup, store Store) {
	svc := New(store)
	rg.POST("/set/:key", svc.setHandler)
	rg.GET("/get/:key", svc.getHandler)
	rg.DELETE("/delete/:key", svc.deleteHandler)
	rg.GET("/has/:key", svc.hasHandler)
}

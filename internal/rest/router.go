package rest

import "github.com/gin-gonic/gin"

func SetupRouter(rg *gin.RouterGroup, svc *Service) {

	rg.POST("/set/:key", svc.SetHandler)
	rg.GET("/get/:key", svc.GetHandler)
	rg.DELETE("/delete/:key", svc.DeleteHandler)
	rg.GET("/has/:key", svc.HasHandler)
}

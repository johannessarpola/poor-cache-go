package rest

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type SetParams struct {
	TTL time.Duration `form:"ttl" binding:"required"`
}

func (s *Service) setHandler(c *gin.Context) {
	key := c.Param("key")
	var value interface{}
	if err := c.ShouldBindJSON(&value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := SetParams{}
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.store.Set(key, value, params.TTL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Value set successfully"})
}

func (s *Service) getHandler(c *gin.Context) {
	key := c.Param("key")
	var dest interface{}
	meta, err := s.store.Get(key, &dest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if meta == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Key not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"meta": meta, "value": dest})
}

func (s *Service) deleteHandler(c *gin.Context) {
	key := c.Param("key")
	if err := s.store.Delete(key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Key deleted successfully"})
}

func (s *Service) hasHandler(c *gin.Context) {
	key := c.Param("key")
	exists := s.store.Has(key)
	c.JSON(http.StatusOK, gin.H{"exists": exists})
}

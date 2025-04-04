package rest

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/johannessarpola/poor-cache-go/internal/logger"
)

var (
	errBadJson  = errors.New("invalid json body")
	errBadQuery = errors.New("invalid query parameters")
	errInternal = errors.New("internal server error")
	errNotFound = errors.New("not found")
)

func newErr(err error) map[string]any {
	return gin.H{"error": err.Error()}
}

type SetParams struct {
	TTL time.Duration `form:"ttl" binding:"required"`
}

func (s *Service) SetHandler(c *gin.Context) {
	key := c.Param("key")
	var value interface{}
	if err := c.ShouldBindJSON(&value); err != nil {
		c.JSON(http.StatusBadRequest, newErr(errBadJson))
		return
	}

	params := SetParams{}
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, newErr(errBadQuery))
		return
	}

	if err := s.store.Set(key, value, params.TTL); err != nil {
		logger.Errorf("Could not set key %s with value %#v", key, value)
		c.JSON(http.StatusInternalServerError, newErr(errInternal))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "success"})
}

func (s *Service) GetHandler(c *gin.Context) {
	key := c.Param("key")
	var dest any
	meta, err := s.store.Get(key, &dest)
	if err != nil {
		logger.Errorf("Could not get key %s", key)
		c.JSON(http.StatusInternalServerError, newErr(errInternal))
		return
	}

	if dest == nil {
		c.JSON(http.StatusNotFound, newErr(errNotFound))
		return
	}

	c.JSON(http.StatusOK, gin.H{"meta": meta, "value": dest})
}

func (s *Service) DeleteHandler(c *gin.Context) {
	key := c.Param("key")
	if err := s.store.Delete(key); err != nil {
		c.JSON(http.StatusInternalServerError, newErr(errInternal))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func (s *Service) HasHandler(c *gin.Context) {
	key := c.Param("key")
	exists := s.store.Has(key)
	c.JSON(http.StatusOK, gin.H{"exists": exists})
}

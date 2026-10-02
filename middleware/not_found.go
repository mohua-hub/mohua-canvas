package middleware

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func NotFoundJSON(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"code": 1, "data": nil, "msg": "接口不存在"})
}

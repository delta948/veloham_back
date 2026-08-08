package payments

import (
	"github.com/gin-gonic/gin"
	"veloham/backend/internal/common"
)

func RegisterRoutes(rg *gin.RouterGroup, deps common.Dependencies) {
	h := NewHandler(deps.DB, deps.Config)
	rg.POST("/payments/freedompay/result", h.Result)
	rg.GET("/payments/quota", deps.AuthMW, h.Quota)
	rg.POST("/payments/:id/checkout", deps.AuthMW, h.Checkout)
	rg.GET("/payments/:id", deps.AuthMW, h.Status)
	admin := rg.Group("/admin/payments", deps.AuthMW, deps.AdminMW)
	admin.GET("", h.AdminList)
	admin.POST("/:id/recheck", h.AdminRecheck)
}

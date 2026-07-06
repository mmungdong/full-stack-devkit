package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/onexstack/onexstack/pkg/core"
)

// ConfigResponse 全局配置响应.
type ConfigResponse struct {
	// DefaultLanguage 默认语言.
	DefaultLanguage string `json:"defaultLanguage"`
}

// GetConfig 返回前端全局配置.
//
// @Summary      获取全局配置
// @Description  返回前端可用的全局配置（如默认语言）
// @Tags         系统
// @Produce      json
// @Success      200  {object}  ConfigResponse
// @Router       /api/config [get]
func (h *Handler) GetConfig(c *gin.Context) {
	core.WriteResponse(c, ConfigResponse{
		DefaultLanguage: h.defaultLanguage,
	}, nil)
}

package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/onexstack/onexstack/pkg/core"
)

// Login 用户登录并返回 JWT Token.
//
// @Summary      用户登录
// @Description  通过用户名密码登录，返回 JWT Token
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body  body      v1.LoginRequest  true  "登录请求"
// @Success      200   {object}  v1.LoginResponse
// @Failure      400   {object}  v1.LoginResponse
// @Router       /api/login [post]
func (h *Handler) Login(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.UserV1().Login, h.val.ValidateLoginRequest)
}

// RefreshToken 刷新 JWT Token.
//
// @Summary      刷新令牌
// @Description  使用当前有效的 Token 刷新获取新 Token
// @Tags         认证
// @Accept       json
// @Produce      json
// @Success      200  {object}  v1.RefreshTokenResponse
// @Router       /api/refresh-token [put]
func (h *Handler) RefreshToken(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.UserV1().RefreshToken)
}

// ChangePassword 修改用户密码.
//
// @Summary      修改密码
// @Description  修改指定用户的密码
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        userID  path      string                  true  "用户 ID"
// @Param        body    body      v1.ChangePasswordRequest  true  "修改密码请求"
// @Success      200     {object}  v1.ChangePasswordResponse
// @Router       /api/v1/users/{userID}/change-password [put]
func (h *Handler) ChangePassword(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.UserV1().ChangePassword, h.val.ValidateChangePasswordRequest)
}

// CreateUser 创建新用户.
//
// @Summary      注册用户
// @Description  创建新用户（注册），无需认证
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        body  body      v1.CreateUserRequest  true  "用户信息"
// @Success      200   {object}  v1.CreateUserResponse
// @Failure      400   {object}  v1.CreateUserResponse
// @Router       /api/v1/users [post]
func (h *Handler) CreateUser(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.UserV1().Create, h.val.ValidateCreateUserRequest)
}

// UpdateUser 更新用户信息.
//
// @Summary      更新用户
// @Description  更新指定用户的信息
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        userID  path      string                  true  "用户 ID"
// @Param        body    body      v1.UpdateUserRequest    true  "更新用户请求"
// @Success      200     {object}  v1.UpdateUserResponse
// @Router       /api/v1/users/{userID} [put]
func (h *Handler) UpdateUser(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.UserV1().Update, h.val.ValidateUpdateUserRequest)
}

// DeleteUser 删除用户.
//
// @Summary      删除用户
// @Description  删除指定用户
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        userID  path      string                  true  "用户 ID"
// @Success      200     {object}  v1.DeleteUserResponse
// @Router       /api/v1/users/{userID} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.UserV1().Delete, h.val.ValidateDeleteUserRequest)
}

// GetUser 获取用户信息.
//
// @Summary      查询用户详情
// @Description  获取指定用户的详细信息
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        userID  path      string                  true  "用户 ID"
// @Success      200     {object}  v1.GetUserResponse
// @Router       /api/v1/users/{userID} [get]
func (h *Handler) GetUser(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.UserV1().Get, h.val.ValidateGetUserRequest)
}

// ListUser 列出用户信息.
//
// @Summary      查询用户列表
// @Description  分页查询用户列表
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        offset  query     int64              false  "偏移量"  default(0)
// @Param        limit   query     int64              false  "每页数量" default(10)
// @Success      200     {object}  v1.ListUserResponse
// @Router       /api/v1/users [get]
func (h *Handler) ListUser(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.UserV1().List, h.val.ValidateListUserRequest)
}

func init() {
	Register(func(v1 *gin.RouterGroup, handler *Handler) {
		// 用户相关路由
		rg := v1.Group("/users")
		rg.POST("", handler.CreateUser) // 创建用户。这里要注意：创建用户是不用进行认证和授权的
		rg.Use(handler.mws...)
		rg.PUT(":userID/change-password", handler.ChangePassword) // 修改用户密码
		rg.PUT(":userID", handler.UpdateUser)                     // 更新用户信息
		rg.DELETE(":userID", handler.DeleteUser)                  // 删除用户
		rg.GET(":userID", handler.GetUser)                        // 查询用户详情
		rg.GET("", handler.ListUser)                              // 查询用户列表.
	})
}

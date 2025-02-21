package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/onexstack/onexstack/pkg/core"
)

// Login 用户登录并返回 JWT Token.
func (h *Handler) Login(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.UserV1().Login, h.val.ValidateLoginRequest)
}

// RefreshToken 刷新 JWT Token.
func (h *Handler) RefreshToken(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.UserV1().RefreshToken)
}

// ChangePassword 修改用户密码.
func (h *Handler) ChangePassword(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.UserV1().ChangePassword, h.val.ValidateChangePasswordRequest)
}

// CreateUser 创建新用户.
func (h *Handler) CreateUser(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.UserV1().Create, h.val.ValidateCreateUserRequest)
}

// UpdateUser 更新用户信息.
func (h *Handler) UpdateUser(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.UserV1().Update, h.val.ValidateUpdateUserRequest)
}

// DeleteUser 删除用户.
func (h *Handler) DeleteUser(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.UserV1().Delete, h.val.ValidateDeleteUserRequest)
}

// GetUser 获取用户信息.
func (h *Handler) GetUser(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.UserV1().Get, h.val.ValidateGetUserRequest)
}

// ListUser 列出用户信息.
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

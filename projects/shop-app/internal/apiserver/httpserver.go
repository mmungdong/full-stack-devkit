package apiserver

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/onexstack/onexstack/pkg/core"
	"github.com/onexstack/onexstack/pkg/server"
	"github.com/onexstack/shop-app/internal/apiserver/handler"
	"github.com/onexstack/shop-app/internal/pkg/errno"
	mw "github.com/onexstack/shop-app/internal/pkg/middleware/gin"
	web "github.com/onexstack/shop-app/frontend"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// 引入 swag 生成的 swagger 文档，注册 SwaggerInfo 供 gin-swagger 读取
	_ "github.com/onexstack/shop-app/docs/apidocs"
)

// ginServer 定义一个使用 Gin 框架开发的 HTTP 服务器.
type ginServer struct {
	srv server.Server
}

// 确保 *ginServer 实现了 server.Server 接口.
var _ server.Server = (*ginServer)(nil)

func (c *ServerConfig) NewGinServer() (*ginServer, error) {
	// 创建 Gin 引擎
	engine := gin.New()

	// 注册全局中间件，用于恢复 panic、设置 HTTP 头、添加请求 ID 等
	engine.Use(
		gin.Recovery(),
		mw.NoCache,
		mw.Cors,
		mw.Secure,
	)

	// 注册.R API 路由
	c.InstallRESTAPI(engine)

	httpsrv := server.NewHTTPServer(c.HTTPOptions, c.TLSOptions, engine)

	return &ginServer{srv: httpsrv}, nil
}

// 注册 API 路由。所有业务接口统一挂在 /api 前缀下，与前端静态文件路由隔离.
func (c *ServerConfig) InstallRESTAPI(engine *gin.Engine) {
	// 注册业务无关的 API 接口（pprof、404、静态文件等，不含 /api）
	InstallGenericAPI(engine)

	// 所有业务 API 统一 /api 前缀
	api := engine.Group("/api")

	// 认证和授权中间件
	authMiddlewares := []gin.HandlerFunc{mw.AuthnMiddleware(c.retriever), mw.AuthzMiddleware(c.authz)}

	// 创建核心业务处理器
	hdl := handler.NewHandler(c.biz, c.val, c.DefaultLanguage, authMiddlewares...)
	// 注册健康检查接口
	api.GET("/healthz", hdl.Healthz)
	// 注册全局配置接口（无需认证，登录页加载时调用）
	api.GET("/config", hdl.GetConfig)
	// 注册用户登录和令牌刷新接口。这2个接口比较简单，所以没有 API 版本
	api.POST("/login", hdl.Login)
	// 注意：认证中间件要在 hdl.RefreshToken 之前加载
	api.PUT("/refresh-token", mw.AuthnMiddleware(c.retriever), hdl.RefreshToken)

	// 注册 v1 版本 API 路由分组（/api/v1）
	v1 := api.Group("/v1")
	// 注册资源路由
	hdl.InstallAll(v1)
}

// InstallGenericAPI 注册业务无关的路由，例如 pprof、静态文件、404 处理等.
func InstallGenericAPI(engine *gin.Engine) {
	// 注册 pprof 路由
	pprof.Register(engine)

	// 注册 Swagger UI 路由，访问 /swagger/index.html 查看全部接口文档
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 托管前端静态文件（go:embed 嵌入的 frontend/out）
	distFS := web.DistFS()
	fileServer := http.FileServer(http.FS(distFS))

	// SPA fallback：非 /api、非 /swagger 的请求交给静态文件服务器
	engine.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// API 路由未命中 → 返回 JSON 404
		if strings.HasPrefix(path, "/api/") {
			core.WriteResponse(c, errno.ErrPageNotFound, nil)
			return
		}
		// 尝试静态文件
		c.Request.URL.Path = path
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}

// RunOrDie 启动 Gin 服务器，出错则程序崩溃退出.
func (s *ginServer) RunOrDie() {
	s.srv.RunOrDie()
}

// GracefulStop 优雅停止服务器.
func (s *ginServer) GracefulStop(ctx context.Context) {
	s.srv.GracefulStop(ctx)
}

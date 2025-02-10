package logs

import (
	"asyncKubeManager/pkg/dbresolver"
	"asyncKubeManager/pkg/server/middleware"
	"asyncKubeManager/pkg/token"
	"github.com/gin-gonic/gin"
)

// RegisterRouter 注册日志相关路由
func RegisterRouter(group *gin.RouterGroup, tokenManager token.Manager, dbResolver *dbresolver.DBResolver) {
	// 初始化日志监听器
	startEventLogListener(dbResolver)
	startUserOperatorLogListener(dbResolver)

	// 创建日志路由组
	logG := group.Group("/logs")

	// 初始化handler
	handler := newLogHandler(logHandlerOption{
		dbResolver: dbResolver,
	})

	// 所有接口都需要token验证
	logG.Use(middleware.CheckToken(tokenManager))

	// 事件日志接口
	logG.POST("/event/list", handler.listEventLogs)
	logG.POST("/event/detail", handler.getEventLog)

	// 用户操作日志接口
	logG.POST("/user/list", handler.listUserOperatorLogs)
	logG.POST("/user/detail", handler.getUserOperatorLog)
}

package logs

import (
	"asyncKubeManager/pkg/dao"
	"asyncKubeManager/pkg/dbresolver"
	"asyncKubeManager/pkg/model"
	"asyncKubeManager/pkg/server/encoding"
	"asyncKubeManager/pkg/server/errutil"
	"asyncKubeManager/pkg/server/request"
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

type logHandlerOption struct {
	dbResolver *dbresolver.DBResolver
}

type logHandler struct {
	logHandlerOption
}

func newLogHandler(option logHandlerOption) *logHandler {
	return &logHandler{
		logHandlerOption: option,
	}
}

// 获取事件日志列表
func (h *logHandler) listEventLogs(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, time.Second*30)
	defer cancel()

	req := listEventLogsReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		encoding.HandleError(c, errutil.ErrJSONFormat)
		return
	}

	if err := request.ValidateStruct(ctx, req); err != nil {
		encoding.HandleError(c, err)
		return
	}

	var logs []model.EventLog
	var err error

	if req.ResourceType != "" {
		logs, err = dao.ListEventLogsByType(ctx, h.dbResolver, model.EventType(req.EventType))
	} else if req.Creator != "" {
		logs, err = dao.ListEventLogsByCreator(ctx, h.dbResolver, req.Creator)
	} else {
		logs, err = dao.ListEventLogs(ctx, h.dbResolver)
	}

	if err != nil {
		zap.L().Error("failed to list event logs", zap.Error(err))
		encoding.HandleError(c, errutil.ErrInternalServer)
		return
	}

	encoding.HandleSuccessList(c, int64(len(logs)), logs)
}

// 获取事件日志详情
func (h *logHandler) getEventLog(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, time.Second*30)
	defer cancel()

	req := getEventLogReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		encoding.HandleError(c, errutil.ErrJSONFormat)
		return
	}

	if err := request.ValidateStruct(ctx, req); err != nil {
		encoding.HandleError(c, err)
		return
	}

	found, log, err := dao.GetEventLogByID(ctx, h.dbResolver, req.ID)
	if err != nil {
		zap.L().Error("failed to get event log", zap.Error(err))
		encoding.HandleError(c, errutil.ErrInternalServer)
		return
	}
	if !found {
		encoding.HandleError(c, errutil.ErrNotFound)
		return
	}

	encoding.HandleSuccess(c, log)
}

// 获取用户操作日志列表
func (h *logHandler) listUserOperatorLogs(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, time.Second*30)
	defer cancel()

	req := listUserOperatorLogsReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		encoding.HandleError(c, errutil.ErrJSONFormat)
		return
	}

	if err := request.ValidateStruct(ctx, req); err != nil {
		encoding.HandleError(c, err)
		return
	}

	var logs []model.UserOperatorLog
	var err error

	if req.UID != "" {
		logs, err = dao.GetUserOperatorLogsByUID(ctx, h.dbResolver, req.UID)
	} else {
		logs, err = dao.ListUserOperatorLogs(ctx, h.dbResolver)
	}

	if err != nil {
		zap.L().Error("failed to list user operator logs", zap.Error(err))
		encoding.HandleError(c, errutil.ErrInternalServer)
		return
	}

	encoding.HandleSuccessList(c, int64(len(logs)), logs)
}

// 获取用户操作日志详情
func (h *logHandler) getUserOperatorLog(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, time.Second*30)
	defer cancel()

	req := getUserOperatorLogReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		encoding.HandleError(c, errutil.ErrJSONFormat)
		return
	}

	if err := request.ValidateStruct(ctx, req); err != nil {
		encoding.HandleError(c, err)
		return
	}

	log, err := dao.GetUserOperatorLogByID(ctx, h.dbResolver, req.ID)
	if err != nil {
		zap.L().Error("failed to get user operator log", zap.Error(err))
		encoding.HandleError(c, errutil.ErrInternalServer)
		return
	}

	encoding.HandleSuccess(c, log)
}

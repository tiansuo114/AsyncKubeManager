package logs

type (
	// 获取事件日志列表请求
	listEventLogsReq struct {
		ResourceType string `json:"resource_type" validate:"omitempty"`
		ResourceUID  string `json:"resource_uid" validate:"omitempty"`
		EventType    string `json:"event_type" validate:"omitempty"`
		Creator      string `json:"creator" validate:"omitempty"`
	}

	// 获取事件日志详情请求
	getEventLogReq struct {
		ID uint `json:"id" validate:"required"`
	}

	// 获取用户操作日志列表请求
	listUserOperatorLogsReq struct {
		UID      string `json:"uid" validate:"omitempty"`
		Operator string `json:"operator" validate:"omitempty"`
	}

	// 获取用户操作日志详情请求
	getUserOperatorLogReq struct {
		ID int64 `json:"id" validate:"required"`
	}
)

package passport

import "asyncKubeManager/pkg/model"

type (
	loginReq struct {
		UserID       string `json:"user_id" validate:"required"`
		Password     string `json:"password" validate:"required"`
		CaptchaID    string `json:"captcha_id" validate:"required"`
		CaptchaValue string `json:"captcha_value" validate:"required"`
	}
	loginResp struct {
		UID      string `json:"uid"`
		Username string `json:"username"`
		Token    string `json:"token"`
	}

	createCaptchaResp struct {
		CaptchaID string `json:"captcha_id"`
		Image     string `json:"image"`
	}

	updateUserReq struct {
		UID      string         `json:"uid" validate:"required"`
		Username string         `json:"username" validate:"required,gt=0,lte=50"`
		Email    string         `json:"email" validate:"email"`
		Tel      string         `json:"tel" validate:"omitempty"`
		Desc     string         `json:"desc" validate:"omitempty"`
		Password string         `json:"password" validate:"omitempty,min=6"` // Optional password field
		Role     model.UserRole `json:"role" validate:"omitempty"`
	}
)

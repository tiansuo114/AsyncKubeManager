package passport

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

	registerLicenseReq struct {
		License string `json:"license"`
	}
)

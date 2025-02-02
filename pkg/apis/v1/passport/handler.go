package passport

import (
	"asyncKubeManager/pkg/captcha"
	"asyncKubeManager/pkg/client/ldap"
	"asyncKubeManager/pkg/dbresolver"
	"asyncKubeManager/pkg/server/encoding"
	"asyncKubeManager/pkg/server/errutil"
	"asyncKubeManager/pkg/server/request"
	"asyncKubeManager/pkg/token"
	"asyncKubeManager/pkg/types"
	"asyncKubeManager/pkg/utils"
	"asyncKubeManager/pkg/utils/limiter"
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type authHandlerOption struct {
	tokenManager   token.Manager
	dbResolver     *dbresolver.DBResolver
	captchaLimiter *limiter.KeyLimiter
	loginLimiter   *limiter.LoginLimiter
	ldapClient     *ldap.LDAPClient
}

type authHandler struct {
	authHandlerOption
}

func newAuthHandler(option authHandlerOption) *authHandler {
	return &authHandler{
		authHandlerOption: option,
	}
}

func (h *authHandler) logout(c *gin.Context) {
	encoding.HandleSuccess(c)
}

func (h *authHandler) createCaptcha(c *gin.Context) {
	if !h.captchaLimiter.AllowKey(utils.MD5Hex(c.Request.UserAgent())) {
		encoding.HandleError(c, errutil.NewError(http.StatusBadRequest, "The captcha request is too fast"))
		return
	}

	captchaId, captchaValue, err := captcha.CreateCaptcha()
	if err != nil {
		zap.L().Error("create captcha failed", zap.Error(err))
		encoding.HandleError(c, errutil.NewError(http.StatusBadRequest, "create captcha failed"))
		return
	}

	encoding.HandleSuccess(c, createCaptchaResp{captchaId, captchaValue})
}

func (h *authHandler) login(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, types.DefaultTimeout)
	defer cancel()

	req := loginReq{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		zap.L().Error("c.ShouldBindJSON", zap.Error(err))
		encoding.HandleError(c, errutil.ErrIllegalParameter)
		return
	}
	req.UserID = strings.TrimSpace(req.UserID)

	if t, err := h.tokenManager.GetTokenFromCtx(c); err == nil {
		if clm, err := h.tokenManager.Verify(t); err == nil {
			encoding.HandleSuccess(c, loginResp{
				UID:      clm.UID,
				Token:    t,
				Username: clm.Username,
			})
			return
		}
	}

	if err = request.ValidateStruct(ctx, req); err != nil {
		encoding.HandleError(c, err)
		return
	}

	if !captcha.VerifyCaptcha(req.CaptchaID, strings.ToLower(req.CaptchaValue)) {
		encoding.HandleError(c, errutil.NewError(http.StatusBadRequest, "captcha value is wrong"))
		return
	}

	ldapUser, err := h.ldapClient.FindUserByUID(req.UserID)
	if err != nil {
		zap.L().Error("FindUserByUID", zap.Error(err))
		encoding.HandleError(c, errutil.NewError(http.StatusBadRequest, "user not found"))
		return
	}

	if err = h.ldapClient.Bind(ldapUser.DN, req.Password); err != nil {
		zap.L().Error("ldap bind failed", zap.Error(err))
		encoding.HandleError(c, errutil.NewError(http.StatusBadRequest, "password is wrong"))
		return
	}

	t, err := h.tokenManager.IssueTo(token.Info{
		UID:      ldapUser.UID,
		Username: ldapUser.CN,
		Name:     ldapUser.CN,
		Primary:  true,
	}, token.DefaultCacheDuration)
	if err != nil {
		zap.L().Error("IssueTo", zap.Error(err))
		encoding.HandleError(c, errutil.NewError(http.StatusInternalServerError, "failed to issue token"))
		return
	}

	// 返回登录成功的响应
	encoding.HandleSuccess(c, loginResp{
		UID:      ldapUser.UID,
		Token:    t,
		Username: ldapUser.CN,
	})
}

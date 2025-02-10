package passport

import (
	"asyncKubeManager/pkg/apis/v1/logs"
	"asyncKubeManager/pkg/captcha"
	"asyncKubeManager/pkg/client/ldap"
	"asyncKubeManager/pkg/dao"
	"asyncKubeManager/pkg/dbresolver"
	"asyncKubeManager/pkg/model"
	"asyncKubeManager/pkg/server/encoding"
	"asyncKubeManager/pkg/server/errutil"
	"asyncKubeManager/pkg/server/request"
	"asyncKubeManager/pkg/token"
	"asyncKubeManager/pkg/types"
	"asyncKubeManager/pkg/utils"
	"asyncKubeManager/pkg/utils/limiter"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"time"
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
	tokenUser := token.GetUIDFromCtx(c)

	req := loginReq{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		zap.L().Error("c.ShouldBindJSON", zap.Error(err))
		encoding.HandleError(c, errutil.ErrIllegalParameter)
		return
	}
	req.UserID = strings.TrimSpace(req.UserID)

	// Check if there is already an active token in the request context
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

	// Validate the incoming login request
	if err = request.ValidateStruct(ctx, req); err != nil {
		encoding.HandleError(c, err)
		return
	}

	// Verify CAPTCHA value
	if !captcha.VerifyCaptcha(req.CaptchaID, strings.ToLower(req.CaptchaValue)) {
		encoding.HandleError(c, errutil.NewError(http.StatusBadRequest, "captcha value is wrong"))
		return
	}

	// Look up the user in LDAP
	ldapUser, err := h.ldapClient.FindUserByUID(req.UserID)
	if err != nil {
		zap.L().Error("FindUserByUID", zap.Error(err))
		encoding.HandleError(c, errutil.NewError(http.StatusBadRequest, "user not found"))
		return
	}

	// Perform LDAP bind to verify user credentials
	if err = h.ldapClient.Bind(ldapUser.DN, req.Password); err != nil {
		zap.L().Error("ldap bind failed", zap.Error(err))
		encoding.HandleError(c, errutil.NewError(http.StatusBadRequest, "password is wrong"))
		return
	}

	// Check if the user already exists in the system's database
	found, user, err := dao.GetUserByUserName(c, h.dbResolver, ldapUser.UID)
	if err != nil {
		zap.L().Error("GetUserByUserName", zap.Error(err))
		encoding.HandleError(c, errutil.ErrInternalServer)
		return
	}

	if found {
		if user != nil {
			t, err := h.tokenManager.IssueTo(token.Info{
				UID:      user.UID,
				Username: user.Username,
				Name:     user.Username,
				Primary:  true,
			}, token.DefaultCacheDuration)
			if err != nil {
				zap.L().Error("IssueTo", zap.Error(err))
				encoding.HandleError(c, errutil.NewError(http.StatusInternalServerError, "failed to issue token"))
				return
			}

			logs.UserOperatorLogChannel <- &model.UserOperatorLog{
				UID:       ldapUser.UID,
				Operator:  model.UserOperatorLogin, // Add the operation type, here it's "user login"
				CreatedAt: time.Now().UnixMilli(),
				Creator:   token.GetUIDFromCtx(c), // Creator of the operation
			}

			// Return the token and user info
			encoding.HandleSuccess(c, loginResp{
				UID:      user.UID,
				Token:    t,
				Username: user.Username,
			})
			return
		}
	} else {
		// Start a new transaction
		err = h.dbResolver.GetDB().Transaction(func(tx *gorm.DB) error {
			// If user doesn't exist, create a new user with LDAP data
			userRole := model.UserRoleNormal
			if strings.Contains(ldapUser.OU, "admin") {
				userRole = model.UserRoleAdmin
			}

			newUser, err := dao.InsertUserWithDB(c, tx, ldapUser.UID, "", ldapUser.TelephoneNumber, ldapUser.Mail, "LDAP User", userRole)
			if err != nil {
				return err
			}

			// Issue a token for the newly created user
			t, err := h.tokenManager.IssueTo(token.Info{
				UID:      newUser.UID,
				Username: newUser.Username,
				Name:     newUser.Username,
				Primary:  true,
			}, token.DefaultCacheDuration)
			if err != nil {
				return err
			}

			logs.UserOperatorLogChannel <- &model.UserOperatorLog{
				UID:       ldapUser.UID,
				Operator:  model.UserOperatorFirstLogin, // Add the operation type, here it's "user login"
				CreatedAt: time.Now().UnixMilli(),
				Creator:   token.GetUIDFromCtx(c), // Creator of the operation
			}

			// Return the token and new user info
			encoding.HandleSuccess(c, loginResp{
				UID:      newUser.UID,
				Token:    t,
				Username: newUser.Username,
			})

			return nil
		})

		if err != nil {
			zap.L().Error("Transaction failed", zap.Error(err))
			encoding.HandleError(c, errutil.ErrInternalServer)
			return
		}
	}
	defer func() {
		if err != nil {
			logs.UserOperatorLogChannel <- &model.UserOperatorLog{
				UID:       req.UserID,
				Operator:  model.UserOperatorError, // Store the JSON string as operator details
				Operation: fmt.Sprintf("user login error : %s", err),
				CreatedAt: time.Now().UnixMilli(),
				Creator:   tokenUser, // Creator of the operation
			}
		}
	}()
}

func (h *authHandler) update(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, types.DefaultTimeout)
	defer cancel()
	var err error
	tokenUser := token.GetUIDFromCtx(c)

	req := updateUserReq{}
	// Binding request body to struct
	if err = c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("", zap.Error(err))
		encoding.HandleError(c, errutil.ErrJSONFormat)
		return
	}
	// Validating the struct based on tags
	if err = request.ValidateStruct(ctx, req); err != nil {
		encoding.HandleError(c, err)
		return
	}

	defer func() {
		if err != nil {
			logs.UserOperatorLogChannel <- &model.UserOperatorLog{
				UID:       req.UID,
				Operator:  model.UserOperatorError, // Store the JSON string as operator details
				Operation: fmt.Sprintf("user update error : %s", err),
				CreatedAt: time.Now().UnixMilli(),
				Creator:   tokenUser, // Creator of the operation
			}
		}
	}()

	// Fetch user by UID
	found, user, err := dao.GetUserByUID(ctx, h.dbResolver, req.UID)
	if err != nil {
		encoding.HandleError(c, err)
		return
	}
	if !found {
		encoding.HandleError(c, errutil.ErrNotFound)
		return
	}

	updated := map[string]interface{}{}
	operatorDetails := map[string]string{}
	if req.Username != "" {
		updated["username"] = req.Username
		operatorDetails["username"] = fmt.Sprintf("username changed from %s to %s", user.Username, req.Username)
	}
	if req.Email != "" {
		updated["email"] = req.Email
		operatorDetails["email"] = fmt.Sprintf("email changed from %s to %s", user.Email, req.Email)
	}
	if req.Tel != "" {
		updated["tel"] = req.Tel
		operatorDetails["tel"] = fmt.Sprintf("tel changed from %s to %s", user.Tel, req.Tel)
	}
	if req.Desc != "" {
		updated["desc"] = req.Desc
		operatorDetails["desc"] = fmt.Sprintf("desc changed from %s to %s", user.Desc, req.Desc)
	}
	if req.Role != "" {
		updated["role"] = req.Role
		operatorDetails["role"] = fmt.Sprintf("role changed from %s to %s", user.Role, req.Role)
	}
	// Convert updated fields to JSON format
	operatorDetailsJson, err := json.Marshal(operatorDetails)
	if err != nil {
		zap.L().Error("failed to marshal operator details", zap.Error(err))
		encoding.HandleError(c, errutil.NewError(http.StatusInternalServerError, "failed to process operator details"))
		return
	}

	// Start a new transaction
	err = h.dbResolver.GetDB().Transaction(func(tx *gorm.DB) error {
		// Update the user in the database
		if err = dao.UpdateUserByUIDWithDB(ctx, tx, req.UID, updated); err != nil {
			return err
		}

		logs.UserOperatorLogChannel <- &model.UserOperatorLog{
			UID:       user.UID,
			Operator:  model.UserOperatorUpdate, // Store the JSON string as operator details
			Operation: string(operatorDetailsJson),
			CreatedAt: time.Now().UnixMilli(),
			Creator:   token.GetUIDFromCtx(c), // Creator of the operation
		}

		// Return success response
		encoding.HandleSuccess(c)

		return nil
	})

	if err != nil {
		zap.L().Error("Transaction failed", zap.Error(err))
		encoding.HandleError(c, errutil.ErrInternalServer)
		return
	}
}

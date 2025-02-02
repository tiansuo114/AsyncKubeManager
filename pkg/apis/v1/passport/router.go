package passport

import (
	"asyncKubeManager/pkg/client/ldap"
	"asyncKubeManager/pkg/dbresolver"
	"asyncKubeManager/pkg/server/middleware"
	"asyncKubeManager/pkg/token"
	"asyncKubeManager/pkg/utils/limiter"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"time"
)

func RegisterRouter(group *gin.RouterGroup, tokenManager token.Manager, dbResolver *dbresolver.DBResolver, ldapClient *ldap.LDAPClient) {
	authG := group.Group("/auth")
	handler := newAuthHandler(authHandlerOption{
		tokenManager:   tokenManager,
		dbResolver:     dbResolver,
		captchaLimiter: limiter.NewKeyLimiter(rate.Every(time.Second), 3),
		loginLimiter:   limiter.NewLoginLimiter(),
		ldapClient:     ldapClient,
	})

	authG.POST("/login", handler.login)
	authG.GET("/captcha", handler.createCaptcha)

	authG.Use(middleware.CheckToken(tokenManager))
	authG.POST("/logout", handler.logout)
}

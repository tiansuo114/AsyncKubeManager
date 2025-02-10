package app

import (
	"asyncKubeManager/cmd/console/app/options"
	"asyncKubeManager/pkg/apis/v1/admin"
	"asyncKubeManager/pkg/apis/v1/disk"
	"asyncKubeManager/pkg/apis/v1/logs"
	"asyncKubeManager/pkg/apis/v1/passport"
	"asyncKubeManager/pkg/apis/v1/vm"
	"asyncKubeManager/pkg/logger"
	"asyncKubeManager/pkg/server"
	"asyncKubeManager/pkg/server/config"
	"asyncKubeManager/pkg/server/middleware"
	"asyncKubeManager/pkg/version/verflag"
	"errors"
	"fmt"
	"go.uber.org/zap"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"net/http"
	"time"

	"context"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/term"
)

func NewAPIServerCommand() *cobra.Command {
	s := options.NewServerRunOptions()
	cmd := &cobra.Command{
		Use:   "console",
		Short: "Launch the async kubevirt manager API server",
		Long: `The Async Kubevirt Manager API server provides REST API endpoints 
for managing virtual machines and related resources.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			verflag.PrintAndExitIfRequested()

			if errs := s.Validate(); len(errs) != 0 {
				return utilerrors.NewAggregate(errs)
			}

			if err := config.ParseConfigFile(s.GenericServerRunOptions.ConfigFilePath); err != nil {
				return err
			}

			s = config.ConfigToServerRunOptions(config.GetGlobalConfig())

			return Run(s, server.SetupSignalHandler())
		},
		SilenceUsage: true,
	}

	fs := cmd.Flags()
	namedFlagSets := s.Flags()

	// 添加全局标志
	globalFlagSet := namedFlagSets.FlagSet("global")
	globalFlagSet.BoolP("help", "h", false, fmt.Sprintf("help for %s", cmd.Name()))

	// 添加所有标志集
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	// 设置使用说明
	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	cliflag.SetUsageAndHelpFunc(cmd, namedFlagSets, cols)

	return cmd
}

func Run(s *options.ServerRunOptions, stopCh <-chan struct{}) error {
	// 创建并初始化服务器
	server, err := NewConsoleServer(s, stopCh)
	if err != nil {
		return err
	}

	// 准备运行
	if err := server.PrepareRun(stopCh); err != nil {
		return err
	}

	// 启动服务器
	return server.Run(stopCh)
}

// PrepareRun 准备运行服务器
func (s *ConsoleServer) PrepareRun(stopCh <-chan struct{}) error {
	gin.DisableBindValidation()
	s.router = gin.New()
	s.router.ContextWithFallback = true
	s.router.Use(gin.Recovery())
	s.router.Use(logger.GinLogger())
	s.router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowCredentials: true,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
	}))

	if err := s.initSystem(); err != nil {
		zap.L().Panic("init system failed", zap.Error(err))
	}

	s.installAPIs()
	s.Server.Handler = s.router

	return nil
}

// Run 运行服务器
func (s *ConsoleServer) Run(stopCh <-chan struct{}) error {
	// 启动HTTP服务器
	go func() {
		if err := s.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.L().Panic("HTTP server ListenAndServe failed", zap.Error(err))
		}
	}()

	// 等待停止信号
	<-stopCh
	zap.L().Info("Shutting down server...")

	// 创建关闭上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 优雅关闭HTTP服务器
	if err := s.Server.Shutdown(ctx); err != nil {
		zap.L().Error("Server forced to shutdown", zap.Error(err))
		return err
	}

	// 关闭数据库连接
	if err := s.DBResolver.Close(); err != nil {
		zap.L().Error("Failed to close database connection", zap.Error(err))
	}

	zap.L().Info("Server exiting")
	return nil
}

func (s *ConsoleServer) installAPIs() {
	apiV1Group := s.router.Group("/api/v1")
	apiV1Group.Use(middleware.AddAuditLog(s.DBResolver))
	admin.RegisterRouter(apiV1Group, s.TokenManager, s.DBResolver)
	disk.RegisterRouter(apiV1Group, s.TokenManager, s.DBResolver, s.PVCManager)
	logs.RegisterRouter(apiV1Group, s.TokenManager, s.DBResolver)
	passport.RegisterRouter(apiV1Group, s.TokenManager, s.DBResolver, s.LDAPClient)
	vm.RegisterRouter(apiV1Group, s.TokenManager, s.DBResolver, s.VMManager)
}

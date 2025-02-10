package app

import (
	"asyncKubeManager/cmd/console/app/options"
	"asyncKubeManager/pkg/auth"
	"asyncKubeManager/pkg/client/cache"
	"asyncKubeManager/pkg/client/k8s"
	"asyncKubeManager/pkg/client/kubevirt"
	"asyncKubeManager/pkg/client/ldap"
	"asyncKubeManager/pkg/dbresolver"
	"asyncKubeManager/pkg/manager/pvc"
	"asyncKubeManager/pkg/manager/vm"
	"asyncKubeManager/pkg/task/delete_task"
	"asyncKubeManager/pkg/token"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"

	cdiCli "kubevirt.io/client-go/containerizeddataimporter"
)

type ConsoleServer struct {
	Server *http.Server
	router *gin.Engine

	// 核心组件
	TokenManager token.Manager
	DBResolver   *dbresolver.DBResolver
	CacheClient  cache.Interface
	Enforcer     *auth.Enforcer

	// 客户端
	K8sClient      *k8s.KubeClient
	KubevirtClient *kubevirt.KubevirtClient
	LDAPClient     *ldap.LDAPClient
	CdiClient      *cdiCli.Clientset

	// manager
	VMManager         *vm.KubevirtVMManager
	PVCManager        *pvc.K8sPVCManager
	DeleteTaskManager deleteTask.DeleteTaskManager

	// 任务管理器
	DeleteTaskMonitor *deleteTask.DeleteTaskMonitor
}

func NewConsoleServer(opts *options.ServerRunOptions, stopCh <-chan struct{}) (*ConsoleServer, error) {
	// 初始化各个组件
	dbResolver, err := dbresolver.NewDBResolver(opts.RDBOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create db resolver: %w", err)
	}

	cacheClient, err := cache.NewRedisClient(opts.CacheOptions, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache client: %w", err)
	}

	k8sClient, err := k8s.NewKubeClient(opts.K8sOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}

	kubevirtClient, err := kubevirt.NewKubevirtClient(opts.KubevirtOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubevirt client: %w", err)
	}

	enforcer, err := auth.NewEnforcer(dbResolver.GetDB())
	if err != nil {
		return nil, fmt.Errorf("failed to create enforcer: %w", err)
	}

	cdiClientSet, err := cdiCli.NewForConfig(k8sClient.GetConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create cdi client: %w", err)
	}

	ldapClient, err := ldap.NewLDAPClient(opts.LDAPOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create ldap client: %w", err)
	}

	pvcManager := pvc.NewK8sPVCManager(k8sClient.GetClientset())

	vmManager := vm.NewKubevirtVMManager(kubevirtClient.GetClientset(), cdiClientSet, dbResolver, pvcManager)

	deleteTaskManager := deleteTask.NewDeleteTaskManager(dbResolver, pvcManager, vmManager)
	deleteTaskMonitor := deleteTask.NewDeleteTaskMonitor(dbResolver, deleteTaskManager)

	server := &ConsoleServer{
		TokenManager: token.NewJWTTokenManager([]byte(opts.JWTSecret), jwt.SigningMethodHS256, token.SetDuration(cacheClient, time.Minute*30)),
		DBResolver:   dbResolver,
		CacheClient:  cacheClient,
		Enforcer:     enforcer,

		K8sClient:      k8sClient,
		KubevirtClient: kubevirtClient,
		LDAPClient:     ldapClient,
		CdiClient:      cdiClientSet,

		VMManager:         vmManager,
		PVCManager:        pvcManager,
		DeleteTaskManager: deleteTaskManager,

		DeleteTaskMonitor: deleteTaskMonitor,
	}

	return server, nil
}

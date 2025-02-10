package options

import (
	"asyncKubeManager/pkg/client/cache"
	"asyncKubeManager/pkg/client/k8s"
	"asyncKubeManager/pkg/client/kubevirt"
	"asyncKubeManager/pkg/client/ldap"
	"asyncKubeManager/pkg/client/mysql"
	"asyncKubeManager/pkg/logger"
	genericoptions "asyncKubeManager/pkg/server/options"

	cliflag "k8s.io/component-base/cli/flag"
)

type ServerRunOptions struct {
	GenericServerRunOptions *genericoptions.ServerRunOptions
	CacheOptions            *cache.Options
	RDBOptions              *mysql.Options
	LoggerOptions           *logger.Options
	K8sOptions              *k8s.Options
	KubevirtOptions         *kubevirt.Options
	LDAPOptions             *ldap.Options

	K8sNameSpace    string
	K8sStorageClass string
	DebugMode       bool
	JWTSecret       string
}

var S ServerRunOptions

const defaultJwtSecret = "cebc04fc7d4383ebf11bf661ba69977e" //MD5 async-km

func NewServerRunOptions() *ServerRunOptions {
	S = ServerRunOptions{
		GenericServerRunOptions: genericoptions.NewServerRunOptions(),
		CacheOptions:            cache.NewRedisOptions(),
		RDBOptions:              mysql.NewMysqlOptions(),
		LoggerOptions:           logger.NewLoggerOptions(),
		K8sOptions:              k8s.NewKubeOptions(),
		KubevirtOptions:         kubevirt.NewKubeOptions(),
		LDAPOptions:             ldap.NewLDAPOptions(),
	}
	return &S
}

func (s *ServerRunOptions) Flags() (fss cliflag.NamedFlagSets) {
	fs := fss.FlagSet("generic")
	fs.BoolVar(&s.DebugMode, "debug", false, "Don't enable this if you don't know what it means.")
	fs.StringVar(&s.K8sNameSpace, "k8s-namespace", "async-km", "The namespace of k8s cluster.")
	fs.StringVar(&s.K8sStorageClass, "k8s-storage-class", "async-km-sc", "The storage class of k8s cluster.")
	fs.StringVar(&s.JWTSecret, "jwt-secret", defaultJwtSecret, "The secret of jet.")
	s.GenericServerRunOptions.AddFlags(fs)
	s.CacheOptions.AddFlags(fss.FlagSet("cache"))
	s.RDBOptions.AddFlags(fss.FlagSet("rdb"))
	s.LoggerOptions.AddFlags(fss.FlagSet("log"))
	s.K8sOptions.AddFlags(fss.FlagSet("k8s"))
	s.KubevirtOptions.AddFlags(fss.FlagSet("kubevirt"))
	s.LDAPOptions.AddFlags(fss.FlagSet("ldap"))

	return fss
}

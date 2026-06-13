package options

import (
	"errors"
	"time"

	genericoptions "github.com/onexstack/onexstack/pkg/options"
	"github.com/spf13/pflag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/mungdong/devkit/internal/apiserver"
)

// ServerOptions contains the configuration options for the server.
type ServerOptions struct {
	// JWTKey 定义 JWT 密钥.
	JWTKey string `json:"jwt-key" mapstructure:"jwt-key"`
	// Expiration 定义 JWT Token 的过期时间.
	Expiration time.Duration `json:"expiration" mapstructure:"expiration"`
	// SecureServingOptions contains the TLS configuration options.
	SecureServingOptions *genericoptions.SecureServingOptions `json:"secure" mapstructure:"secure"`
	// InsecureServingOptions contains the HTTP configuration options.
	InsecureServingOptions *genericoptions.InsecureServingOptions `json:"insecure" mapstructure:"insecure"`
	// MySQLOptions contains the MySQL configuration options.
	MySQLOptions *genericoptions.MySQLOptions `json:"coredb" mapstructure:"coredb"`
	// OTelOptions used to specify the otel options.
	OTelOptions *genericoptions.OTelOptions `json:"otel" mapstructure:"otel"`
}

// NewServerOptions creates a ServerOptions instance with default values.
func NewServerOptions() *ServerOptions {
	opts := &ServerOptions{
		JWTKey:                 "",
		Expiration:             2 * time.Hour,
		SecureServingOptions:   genericoptions.NewSecureServingOptions(),
		InsecureServingOptions: genericoptions.NewInsecureServingOptions(),
		MySQLOptions:           genericoptions.NewMySQLOptions(),
		OTelOptions:            genericoptions.NewOTelOptions(),
	}
	opts.InsecureServingOptions.Addr = ":5555"

	return opts
}

// AddFlags binds the options in ServerOptions to command-line flags.
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.JWTKey, "jwt-key", o.JWTKey, "JWT signing key. Must be at least 6 characters long.")
	// 绑定 JWT Token 的过期时间选项到命令行标志。
	// 参数名称为 `--expiration`，默认值为 o.Expiration
	fs.DurationVar(&o.Expiration, "expiration", o.Expiration, "The expiration duration of JWT tokens.")
	// Add command-line flags for sub-options.
	o.SecureServingOptions.AddFlags(fs, "secure")
	o.InsecureServingOptions.AddFlags(fs, "insecure")
	o.MySQLOptions.AddFlags(fs, "coredb")
	o.OTelOptions.AddFlags(fs, "otel")
}

// Complete completes all the required options.
func (o *ServerOptions) Complete() error {
	// TODO: Add the completion logic if needed.
	return nil
}

// Validate checks whether the options in ServerOptions are valid.
func (o *ServerOptions) Validate() error {
	errs := []error{}
	// 校验 JWTKey 长度
	if len(o.JWTKey) < 6 {
		errs = append(errs, errors.New("JWTKey must be at least 6 characters long"))
	}

	// Validate sub-options.
	errs = append(errs, o.SecureServingOptions.Validate()...)
	errs = append(errs, o.InsecureServingOptions.Validate()...)
	errs = append(errs, o.MySQLOptions.Validate()...)
	errs = append(errs, o.OTelOptions.Validate()...)

	// Aggregate all errors and return them.
	return utilerrors.NewAggregate(errs)
}

// Config builds an apiserver.Config based on ServerOptions.
func (o *ServerOptions) Config() (*apiserver.Config, error) {
	return &apiserver.Config{
		JWTKey:                 o.JWTKey,
		Expiration:             o.Expiration,
		SecureServingOptions:   o.SecureServingOptions,
		InsecureServingOptions: o.InsecureServingOptions,
		MySQLOptions:           o.MySQLOptions,
	}, nil
}

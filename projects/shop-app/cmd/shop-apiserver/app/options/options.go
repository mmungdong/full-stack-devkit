// nolint: err113
package options

import (
	"errors"
	"time"

	genericoptions "github.com/onexstack/onexstack/pkg/options"
	"github.com/spf13/pflag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/onexstack/shop-app/internal/apiserver"
)

// ServerOptions contains the configuration options for the server.
type ServerOptions struct {
	// JWTKey 定义 JWT 密钥.
	JWTKey string `json:"jwt-key" mapstructure:"jwt-key"`
	// Expiration 定义 JWT Token 的过期时间.
	Expiration time.Duration `json:"expiration" mapstructure:"expiration"`
	// TLSOptions contains the TLS configuration options.
	TLSOptions *genericoptions.TLSOptions `json:"tls" mapstructure:"tls"`
	// HTTPOptions contains the HTTP configuration options.
	HTTPOptions *genericoptions.HTTPOptions `json:"http" mapstructure:"http"`
	// MySQLOptions contains the MySQL configuration options.
	MySQLOptions *genericoptions.MySQLOptions `json:"mysql" mapstructure:"mysql"`
	// SlogOptions used to specify the slog options.
	SlogOptions *genericoptions.SlogOptions `json:"slog" mapstructure:"slog"`
	// I18nOptions 包含国际化配置.
	I18nOptions *I18nOptions `json:"i18n" mapstructure:"i18n"`
}

// NewServerOptions creates a ServerOptions instance with default values.
func NewServerOptions() *ServerOptions {
	opts := &ServerOptions{
		JWTKey:       "",
		Expiration:   2 * time.Hour,
		TLSOptions:   genericoptions.NewTLSOptions(),
		HTTPOptions:  genericoptions.NewHTTPOptions(),
		MySQLOptions: genericoptions.NewMySQLOptions(),
		SlogOptions:  genericoptions.NewSlogOptions(),
		I18nOptions:  NewI18nOptions(),
	}
	opts.HTTPOptions.Addr = ":5555"

	return opts
}

// AddFlags binds the options in ServerOptions to command-line flags.
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.JWTKey, "jwt-key", o.JWTKey, "JWT signing key. Must be at least 6 characters long.")
	// 绑定 JWT Token 的过期时间选项到命令行标志。
	// 参数名称为 `--expiration`，默认值为 o.Expiration
	fs.DurationVar(&o.Expiration, "expiration", o.Expiration, "The expiration duration of JWT tokens.")
	// Add command-line flags for sub-options.
	o.TLSOptions.AddFlags(fs, "tls")
	o.HTTPOptions.AddFlags(fs, "http")
	o.MySQLOptions.AddFlags(fs, "mysql")
	o.SlogOptions.AddFlags(fs, "slog")
	o.I18nOptions.AddFlags(fs)
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
	errs = append(errs, o.TLSOptions.Validate()...)
	errs = append(errs, o.HTTPOptions.Validate()...)
	errs = append(errs, o.MySQLOptions.Validate()...)

	errs = append(errs, o.SlogOptions.Validate()...)

	// Aggregate all errors and return them.
	return utilerrors.NewAggregate(errs)
}

// Config builds an apiserver.Config based on ServerOptions.
func (o *ServerOptions) Config() (*apiserver.Config, error) {
	return &apiserver.Config{
		JWTKey:       o.JWTKey,
		Expiration:   o.Expiration,
		TLSOptions:   o.TLSOptions,
		HTTPOptions:  o.HTTPOptions,
		MySQLOptions: o.MySQLOptions,
		DefaultLanguage: o.I18nOptions.DefaultLanguage,
	}, nil
}

// I18nOptions 包含国际化相关配置.
type I18nOptions struct {
	// DefaultLanguage 定义默认语言（如 zh、en）.
	DefaultLanguage string `json:"defaultLanguage" mapstructure:"defaultLanguage"`
}

// NewI18nOptions 创建带默认值的 I18nOptions.
func NewI18nOptions() *I18nOptions {
	return &I18nOptions{
		DefaultLanguage: "zh",
	}
}

// AddFlags 绑定 i18n 选项到命令行标志.
func (o *I18nOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.DefaultLanguage, "i18n.default-language", o.DefaultLanguage, "Default language for i18n (e.g. zh, en).")
}

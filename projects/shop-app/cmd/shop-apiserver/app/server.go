package app

import (
	"context"
	"fmt"

	"github.com/onexstack/onexstack/pkg/cli/cli"
	"github.com/onexstack/onexstack/pkg/core"
	"github.com/onexstack/onexstack/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	genericapiserver "k8s.io/apiserver/pkg/server"

	"github.com/onexstack/shop-app/cmd/shop-apiserver/app/options"
)

const (
	// defaultHomeDir defines the default directory to store the configuration for the shop-apiserver service.
	defaultHomeDir = ".shop-app"

	// defaultConfigName specifies the default configuration file name for the shop-apiserver service.
	defaultConfigName = "shop-apiserver.yaml"
)

// Path to the configuration file
var configFile string

// NewWebServerCommand creates a *cobra.Command object used to start the application.
func NewWebServerCommand() *cobra.Command {
	// Create default application command-line options
	opts := options.NewServerOptions()

	cmd := &cobra.Command{
		// Specify the name of the command, which will appear in the help information
		Use: "shop-apiserver",
		// A short description of the command
		Short: "Shop App - 电商后端服务（用户登录注册 + MariaDB 存储）",
		// A detailed description of the command
		Long: `基于 OneX 技术栈脚手架生成的电商后端项目，含用户登录注册模块。缓存层（Redis）后续迭代补充。`,
		// Do not print help information when the command encounters an error.
		// Setting this to true ensures that errors are immediately visible.
		SilenceUsage: true,
		// Specify the Run function to execute when cmd.Execute() is called
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := genericapiserver.SetupSignalContext()

			// If the --version flag is passed, print version information and exit
			version.PrintAndExitIfRequested()

			// Unmarshal the configuration from viper into opts
			if err := viper.Unmarshal(opts); err != nil {
				return fmt.Errorf("failed to unmarshal configuration: %w", err)
			}

			// Complete the options by setting default values and derived configurations
			if err := opts.Complete(); err != nil {
				return fmt.Errorf("failed to complete options: %w", err)
			}

			// Validate command-line options
			if err := opts.Validate(); err != nil {
				return fmt.Errorf("invalid options: %w", err)
			}
			if err := opts.SlogOptions.Apply(); err != nil {
				return err
			}

			return run(ctx, opts)
		},
		// Set argument validation for the command. No command-line arguments are required.
		// For example: ./shop-apiserver param1 param2
		Args: cobra.NoArgs,
	}

	// Initialize configuration function, called when each command runs
	cobra.OnInitialize(core.OnInitialize(&configFile, "SHOP-APP_APISERVER", cli.SearchDirs(defaultHomeDir), defaultConfigName))

	// cobra supports persistent flags, which apply to the assigned command and all its subcommands.
	// It is recommended to use configuration files for application configuration to make it easier to manage configuration items.
	cmd.PersistentFlags().StringVarP(
		&configFile,
		"config",
		"c",
		cli.FilePath(defaultHomeDir, defaultConfigName),
		"Path to the shop-apiserver configuration file.",
	)

	// Add server options as flags
	opts.AddFlags(cmd.PersistentFlags())

	// Add the --version flag
	version.AddFlags(cmd.PersistentFlags())

	return cmd
}

// run contains the main logic for initializing and running the server.
func run(ctx context.Context, opts *options.ServerOptions) error {
	// Retrieve application configuration
	// Separating command-line options and application configuration allows more flexible handling of these two types of configurations.
	cfg, err := opts.Config()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create and start the server
	server, err := cfg.NewServer(ctx)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Run the server
	return server.Run(ctx)
}

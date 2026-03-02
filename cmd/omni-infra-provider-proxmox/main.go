// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package main is the root cmd of the provider script.
package main

import (
	"context"
	"crypto/tls"
	_ "embed"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	dockerclient "github.com/docker/docker/client"
	"github.com/luthermonson/go-proxmox"
	"github.com/siderolabs/omni/client/pkg/client"
	"github.com/siderolabs/omni/client/pkg/infra"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.yaml.in/yaml/v4"

	"github.com/siderolabs/omni-infra-provider-proxmox/internal/pkg/config"
	"github.com/siderolabs/omni-infra-provider-proxmox/internal/pkg/provider"
	localprovider "github.com/siderolabs/omni-infra-provider-proxmox/internal/pkg/provider/local"
	"github.com/siderolabs/omni-infra-provider-proxmox/internal/pkg/provider/meta"
)

//go:embed data/schema.json
var schema string

//go:embed data/icon.svg
var icon []byte

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:          "provider",
	Short:        "Proxmox Omni infrastructure provider",
	Long:         `Connects to Omni as an infra provider and manages VMs in Proxmox`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, _ []string) error {
		loggerConfig := zap.NewProductionConfig()

		logger, err := loggerConfig.Build(
			zap.AddStacktrace(zapcore.ErrorLevel),
		)
		if err != nil {
			return fmt.Errorf("failed to create logger: %w", err)
		}

		if cfg.localMode {
			return runLocalMode(cmd.Context(), logger)
		}

		if !cfg.dryRun && cfg.omniAPIEndpoint == "" {
			return fmt.Errorf("omni-api-endpoint flag is not set")
		}

		var proxmoxConfig config.Config

		configFile, err := os.Open(cfg.configFile)
		if err != nil {
			return fmt.Errorf("failed to read Proxmox config file %q", cfg.configFile)
		}

		decoder := yaml.NewDecoder(configFile)

		if err = decoder.Decode(&proxmoxConfig); err != nil {
			return fmt.Errorf("failed to read Proxmox config file %q", cfg.configFile)
		}

		var opts []proxmox.Option

		switch {
		case proxmoxConfig.Proxmox.Password != "" && proxmoxConfig.Proxmox.Username != "":
			opts = append(opts, proxmox.WithCredentials(&proxmox.Credentials{
				Username: proxmoxConfig.Proxmox.Username,
				Password: proxmoxConfig.Proxmox.Password,
				Realm:    proxmoxConfig.Proxmox.Realm,
			}))
		case proxmoxConfig.Proxmox.TokenID != "" && proxmoxConfig.Proxmox.TokenSecret != "":
			opts = append(opts, proxmox.WithAPIToken(proxmoxConfig.Proxmox.TokenID, proxmoxConfig.Proxmox.TokenSecret))
		}

		if proxmoxConfig.Proxmox.InsecureSkipVerify {
			httpClient := &http.Client{
				Timeout: time.Second * 30,
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			}

			logger.Info("using insecure connection to Proxmox")

			opts = append(opts, proxmox.WithHTTPClient(
				httpClient,
			))
		}

		proxmoxClient := proxmox.NewClient(
			proxmoxConfig.Proxmox.URL,
			opts...,
		)

		if cfg.dryRun {
			return runDryRun(cmd.Context(), logger, proxmoxClient, proxmoxConfig)
		}

		provisioner := provider.NewProvisioner(proxmoxClient)

		ip, err := infra.NewProvider(meta.ProviderID, provisioner, infra.ProviderConfig{
			Name:        cfg.providerName,
			Description: cfg.providerDescription,
			Icon:        base64.RawStdEncoding.EncodeToString(icon),
			Schema:      schema,
		})
		if err != nil {
			return fmt.Errorf("failed to create infra provider: %w", err)
		}

		logger.Info("starting infra provider")

		clientOptions := []client.Option{
			client.WithInsecureSkipTLSVerify(cfg.insecureSkipVerify),
		}

		if cfg.serviceAccountKey != "" {
			clientOptions = append(clientOptions, client.WithServiceAccount(cfg.serviceAccountKey))
		}

		return ip.Run(cmd.Context(), logger, infra.WithOmniEndpoint(cfg.omniAPIEndpoint), infra.WithClientOptions(
			clientOptions...,
		), infra.WithEncodeRequestIDsIntoTokens())
	},
}

var cfg struct {
	omniAPIEndpoint     string
	serviceAccountKey   string
	providerName        string
	providerDescription string
	configFile          string
	insecureSkipVerify  bool
	dryRun              bool
	localMode           bool
}

func main() {
	if err := app(); err != nil {
		os.Exit(1)
	}
}

func app() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)
	defer cancel()

	return rootCmd.ExecuteContext(ctx)
}

func init() {
	rootCmd.Flags().StringVar(&cfg.omniAPIEndpoint, "omni-api-endpoint", os.Getenv("OMNI_ENDPOINT"),
		"the endpoint of the Omni API, if not set, defaults to OMNI_ENDPOINT env var.")
	rootCmd.Flags().StringVar(&meta.ProviderID, "id", meta.ProviderID, "the id of the infra provider, it is used to match the resources with the infra provider label.")
	rootCmd.Flags().StringVar(&cfg.serviceAccountKey, "omni-service-account-key", os.Getenv("OMNI_SERVICE_ACCOUNT_KEY"), "Omni service account key, if not set, defaults to OMNI_SERVICE_ACCOUNT_KEY.")
	rootCmd.Flags().StringVar(&cfg.providerName, "provider-name", "Proxmox", "provider name as it appears in Omni")
	rootCmd.Flags().StringVar(&cfg.providerDescription, "provider-description", "Proxmox infrastructure provider", "Provider description as it appears in Omni")
	rootCmd.Flags().BoolVar(&cfg.insecureSkipVerify, "insecure-skip-verify", false, "ignores untrusted certs on Omni side")
	rootCmd.Flags().BoolVar(&cfg.dryRun, "dry-run", false, "validate config and test Proxmox connectivity without connecting to Omni")
	rootCmd.Flags().BoolVar(&cfg.localMode, "local", false, "use local Docker backend instead of Proxmox (for development)")

	// Read everything into this config file
	rootCmd.Flags().StringVar(&cfg.configFile, "config-file", "", "Proxmox provider config")
}

func runDryRun(ctx context.Context, logger *zap.Logger, proxmoxClient *proxmox.Client, proxmoxConfig config.Config) error {
	logger.Info("dry-run mode: validating configuration and Proxmox connectivity")
	logger.Info("proxmox config", zap.String("url", proxmoxConfig.Proxmox.URL))

	version, err := proxmoxClient.Version(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to Proxmox API: %w", err)
	}

	logger.Info("proxmox connection successful", zap.String("version", version.Version), zap.String("release", version.Release))

	nodes, err := proxmoxClient.Nodes(ctx)
	if err != nil {
		return fmt.Errorf("failed to list Proxmox nodes: %w", err)
	}

	for _, node := range nodes {
		logger.Info("found proxmox node",
			zap.String("name", node.Node),
			zap.String("status", node.Status),
			zap.Uint64("maxMem", node.MaxMem),
			zap.Uint64("mem", node.Mem),
			zap.Int("maxCPU", node.MaxCPU),
		)
	}

	logger.Info("dry-run complete: all checks passed", zap.Int("nodes", len(nodes)))

	return nil
}

func runLocalMode(ctx context.Context, logger *zap.Logger) error {
	logger.Info("starting in local Docker mode")

	docker, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer docker.Close() //nolint:errcheck

	// Verify Docker connectivity
	info, err := docker.Info(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to Docker: %w", err)
	}

	logger.Info("connected to Docker",
		zap.String("serverVersion", info.ServerVersion),
		zap.Int("containers", info.Containers),
		zap.String("name", info.Name),
	)

	provisioner := localprovider.NewProvisioner(docker)

	ip, err := infra.NewProvider(meta.ProviderID, provisioner, infra.ProviderConfig{
		Name:        cfg.providerName,
		Description: cfg.providerDescription + " (local Docker mode)",
		Icon:        base64.RawStdEncoding.EncodeToString(icon),
		Schema:      schema,
	})
	if err != nil {
		return fmt.Errorf("failed to create infra provider: %w", err)
	}

	logger.Info("starting infra provider in local mode")

	clientOptions := []client.Option{
		client.WithInsecureSkipTLSVerify(cfg.insecureSkipVerify),
	}

	if cfg.serviceAccountKey != "" {
		clientOptions = append(clientOptions, client.WithServiceAccount(cfg.serviceAccountKey))
	}

	return ip.Run(ctx, logger, infra.WithOmniEndpoint(cfg.omniAPIEndpoint), infra.WithClientOptions(
		clientOptions...,
	), infra.WithEncodeRequestIDsIntoTokens())
}

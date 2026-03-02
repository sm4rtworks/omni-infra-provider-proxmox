// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package local implements a local Docker-based provisioner for development and testing.
package local

import (
	"context"
	"fmt"
	"strings"
	"time"

	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	dockerclient "github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/siderolabs/omni/client/pkg/infra/provision"
	"github.com/siderolabs/omni/client/pkg/omni/resources/infra"
	"go.uber.org/zap"

	"github.com/siderolabs/omni-infra-provider-proxmox/internal/pkg/provider/resources"
)

const (
	// LabelPrefix is used to tag containers managed by this provisioner.
	LabelPrefix = "omni-infra-provider-proxmox"
	// DefaultImage is the container image used to simulate machines.
	DefaultImage = "alpine:3.21"
)

// Provisioner implements a local Docker-based infra provider for development.
type Provisioner struct {
	docker *dockerclient.Client
}

// NewProvisioner creates a new local Docker provisioner.
func NewProvisioner(docker *dockerclient.Client) *Provisioner {
	return &Provisioner{
		docker: docker,
	}
}

// ProvisionSteps implements infra.Provisioner.
func (p *Provisioner) ProvisionSteps() []provision.Step[*resources.Machine] {
	return []provision.Step[*resources.Machine]{
		provision.NewStep("assignUUID", func(_ context.Context, logger *zap.Logger, pctx provision.Context[*resources.Machine]) error {
			if pctx.State.TypedSpec().Value.Uuid == "" {
				pctx.State.TypedSpec().Value.Uuid = uuid.NewString()
				pctx.SetMachineUUID(pctx.State.TypedSpec().Value.Uuid)
			}

			pctx.State.TypedSpec().Value.Node = "local-docker"

			logger.Info("assigned UUID for local machine",
				zap.String("uuid", pctx.State.TypedSpec().Value.Uuid),
			)

			return nil
		}),
		provision.NewStep("createSchematic", func(ctx context.Context, logger *zap.Logger, pctx provision.Context[*resources.Machine]) error {
			schematic, err := pctx.GenerateSchematicID(ctx, logger,
				provision.WithExtraExtensions("siderolabs/qemu-guest-agent"),
				provision.WithoutConnectionParams(),
			)
			if err != nil {
				return err
			}

			pctx.State.TypedSpec().Value.Schematic = schematic

			return nil
		}),
		provision.NewStep("pullImage", func(ctx context.Context, logger *zap.Logger, pctx provision.Context[*resources.Machine]) error {
			logger.Info("pulling container image", zap.String("image", DefaultImage))

			reader, err := p.docker.ImagePull(ctx, DefaultImage, image.PullOptions{})
			if err != nil {
				return fmt.Errorf("failed to pull image: %w", err)
			}
			defer reader.Close() //nolint:errcheck

			// Drain the reader to complete the pull
			buf := make([]byte, 4096)
			for {
				_, readErr := reader.Read(buf)
				if readErr != nil {
					break
				}
			}

			pctx.State.TypedSpec().Value.TalosVersion = pctx.GetTalosVersion()

			return nil
		}),
		provision.NewStep("createContainer", func(ctx context.Context, logger *zap.Logger, pctx provision.Context[*resources.Machine]) error {
			return p.createContainer(ctx, logger, pctx)
		}),
		provision.NewStep("startContainer", func(ctx context.Context, logger *zap.Logger, pctx provision.Context[*resources.Machine]) error {
			machineUUID := pctx.State.TypedSpec().Value.Uuid

			containerID, err := p.findContainer(ctx, machineUUID)
			if err != nil {
				return err
			}

			if containerID == "" {
				return fmt.Errorf("container for machine %s not found", machineUUID)
			}

			inspect, inspectErr := p.docker.ContainerInspect(ctx, containerID)
			if inspectErr != nil {
				return fmt.Errorf("failed to inspect container: %w", inspectErr)
			}

			if inspect.State.Running {
				logger.Info("container already running", zap.String("containerID", containerID[:12]))

				return nil
			}

			if err := p.docker.ContainerStart(ctx, containerID, dockercontainer.StartOptions{}); err != nil {
				return fmt.Errorf("failed to start container: %w", err)
			}

			logger.Info("started local container", zap.String("containerID", containerID[:12]))

			// Give it a moment to initialize
			time.Sleep(time.Second)

			return nil //nolint:nlreturn
		}),
	}
}

// Deprovision implements infra.Provisioner.
func (p *Provisioner) Deprovision(ctx context.Context, logger *zap.Logger, machine *resources.Machine, _ *infra.MachineRequest) error {
	machineUUID := machine.TypedSpec().Value.Uuid
	if machineUUID == "" {
		return nil
	}

	containerID, err := p.findContainer(ctx, machineUUID)
	if err != nil {
		return err
	}

	if containerID == "" {
		logger.Info("container not found, already removed", zap.String("uuid", machineUUID))

		return nil
	}

	timeout := 10
	if err := p.docker.ContainerStop(ctx, containerID, dockercontainer.StopOptions{Timeout: &timeout}); err != nil {
		if !strings.Contains(err.Error(), "No such container") {
			logger.Warn("failed to stop container", zap.Error(err))
		}
	}

	if err := p.docker.ContainerRemove(ctx, containerID, dockercontainer.RemoveOptions{Force: true}); err != nil {
		if !strings.Contains(err.Error(), "No such container") {
			return fmt.Errorf("failed to remove container: %w", err)
		}
	}

	logger.Info("deprovisioned local container", zap.String("containerID", containerID[:12]))

	return nil
}

// findContainer looks up a container by its machine UUID label.
func (p *Provisioner) findContainer(ctx context.Context, machineUUID string) (string, error) {
	containers, err := p.docker.ContainerList(ctx, dockercontainer.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", LabelPrefix+"/uuid="+machineUUID),
		),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list containers: %w", err)
	}

	if len(containers) == 0 {
		return "", nil
	}

	return containers[0].ID, nil
}

// localData mirrors the provider Data struct for local mode.
type localData struct {
	Memory int64 `yaml:"memory"`
	Cores  int   `yaml:"cores"`
}

// createContainer handles the container creation step.
func (p *Provisioner) createContainer(ctx context.Context, logger *zap.Logger, pctx provision.Context[*resources.Machine]) error {
	machineUUID := pctx.State.TypedSpec().Value.Uuid
	requestID := pctx.GetRequestID()

	// Check if container already exists
	existing, err := p.findContainer(ctx, machineUUID)
	if err != nil {
		return err
	}

	if existing != "" {
		logger.Info("container already exists", zap.String("containerID", existing))

		return nil
	}

	var data localData

	if unmarshalErr := pctx.UnmarshalProviderData(&data); unmarshalErr != nil {
		// Use defaults if no data
		data = localData{
			Memory: 512,
			Cores:  2,
		}
	}

	if data.Memory == 0 {
		data.Memory = 512
	}

	if data.Cores == 0 {
		data.Cores = 2
	}

	containerName := fmt.Sprintf("omni-machine-%s", requestID)

	containerConfig := &dockercontainer.Config{
		Image:    DefaultImage,
		Hostname: requestID,
		Labels: map[string]string{
			LabelPrefix + "/managed":    "true",
			LabelPrefix + "/uuid":       machineUUID,
			LabelPrefix + "/request-id": requestID,
		},
		// Keep container running with a sleep loop
		Cmd: []string{"sh", "-c", "while true; do sleep 3600; done"},
	}

	// Set memory limit in bytes
	memoryBytes := data.Memory * 1024 * 1024

	hostConfig := &dockercontainer.HostConfig{
		Resources: dockercontainer.Resources{
			Memory:   memoryBytes,
			NanoCPUs: int64(data.Cores) * 1e9,
		},
	}

	resp, err := p.docker.ContainerCreate(ctx, containerConfig, hostConfig, &network.NetworkingConfig{}, nil, containerName)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	logger.Info("created local container",
		zap.String("containerID", resp.ID[:12]),
		zap.String("name", containerName),
		zap.Int64("memoryMB", data.Memory),
		zap.Int("cores", data.Cores),
	)

	return nil
}

package task

import (
	"context"
	container2 "github.com/docker/docker/api/types/container"
	image2 "github.com/docker/docker/api/types/image"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/moby/moby/client"
	"github.com/moby/moby/pkg/stdcopy"
	"io"
	"log"
	"math"
	"os"
	"time"
)

type State int

const (
	Pending State = iota
	Scheduled
	Running
	Completed
	Failed
)

type Config struct {
	Name          string
	AttachStdin   bool
	AttachStdout  bool
	AttachStderr  bool
	ExposedPorts  nat.PortSet
	Cmd           []string
	Image         string
	Cpu           float64
	Memory        int64
	Disk          int64
	Env           []string
	RestartPolicy string
	Runtime       DockerRuntime
}

type Docker struct {
	Client *client.Client
	Config Config
}

type DockerRuntime struct {
	ContainerID string
}

type DockerResult struct {
	Error       error
	Action      string
	ContainerId string
	Result      string
}

type Task struct {
	ID            uuid.UUID
	ContainerId   string
	Cpu           float64
	Name          string
	State         State
	Image         string
	Memory        int64
	Disk          int64
	ExposedPorts  nat.PortSet
	PortBindings  map[string]string
	RestartPolicy string
	StartTime     time.Time
	FinishTime    time.Time
}

type TaskEvent struct {
	ID        uuid.UUID
	State     State
	Timestamp time.Time
	Task      Task
}

func (d *Docker) Run() DockerResult {
	ctx := context.Background()
	reader, err := d.Client.ImagePull(
		ctx, d.Config.Image, image2.PullOptions{})
	if err != nil {
		log.Printf("Error pulling image %s: %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}

	_, cErr := io.Copy(os.Stdout, reader)
	if cErr != nil {
		return DockerResult{}
	}

	rp := container2.RestartPolicy{
		Name: container2.RestartPolicyMode(d.Config.RestartPolicy),
	}

	r := container2.Resources{
		Memory:   d.Config.Memory,
		NanoCPUs: int64(d.Config.Cpu * math.Pow(10, 9)),
	}

	cc := container2.Config{
		Image:        d.Config.Image,
		Tty:          false,
		Env:          d.Config.Env,
		ExposedPorts: d.Config.ExposedPorts,
	}

	hc := container2.HostConfig{
		RestartPolicy:   rp,
		Resources:       r,
		PublishAllPorts: true,
	}

	resp, err := d.Client.ContainerCreate(ctx, &cc, &hc, nil, nil, d.Config.Name)
	if err != nil {
		log.Printf("Error creating container using image %s: %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}

	d.Config.Runtime.ContainerID = resp.ID

	out, err := d.Client.ContainerLogs(
		ctx,
		resp.ID,
		container2.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
		},
	)
	if err != nil {
		log.Printf("Error getting logs for container %s: %v\n", resp.ID, err)
		return DockerResult{Error: err}
	}

	_, coErr := stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	if coErr != nil {
		return DockerResult{}
	}

	return DockerResult{ContainerId: resp.ID, Action: "start", Result: "success"}
}

func (d *Docker) Stop(id string) DockerResult {
	log.Printf("Attempting to stop container %s\n", id)
	ctx := context.Background()
	err := d.Client.ContainerStop(ctx, id, container2.StopOptions{})
	if err != nil {
		log.Printf("Error stopping container %s: %v\n", id, err)
		return DockerResult{Error: err}
	}

	rErr := d.Client.ContainerRemove(ctx, id, container2.RemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   false,
		Force:         false,
	})
	if rErr != nil {
		log.Printf("Error removing container %s: %v\n", id, rErr)
		return DockerResult{Error: rErr}
	}

	return DockerResult{Action: "stop", ContainerId: id, Result: "success", Error: nil}
}

func NewConfig(t Task) Config {
	return Config{
		Name:          t.Name,
		ExposedPorts:  t.ExposedPorts,
		Image:         t.Image,
		Cpu:           t.Cpu,
		Memory:        t.Memory,
		Disk:          t.Disk,
		RestartPolicy: "always",
	}
}

func NewDocker(c Config) Docker {
	dc, _ := client.NewClientWithOpts(client.FromEnv)
	return Docker{
		Client: dc,
		Config: c,
	}
}

/*func (cli *Client) ContainerCreate(ctx context.Context, config *container2.Config, hostConfig *container2.HostConfig, networkingConfig *network.NetworkingConfig, platform *specs.Platform, containerName string) (container.ContainerCreateCreatedBody, error) {
	return client.ContainerCreate(ctx, config, hostConfig, networkingConfig, platform, containerName)
}*/

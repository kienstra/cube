package main

import (
	"cube/task"
	"cube/worker"
	"fmt"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/moby/moby/client"
	"time"
)

func main() {
	db := make(map[uuid.UUID]*task.Task)
	w := worker.Worker{
		Queue: *queue.New(),
		Db:    db,
	}

	t := task.Task{
		ID:    uuid.New(),
		Name:  "test-container-1",
		State: task.Scheduled,
		Image: "strm:helloworld-http",
	}

	fmt.Println("starting task")
	w.AddTask(t)
	result := w.RunTask()
	if result.Error != nil {
		panic(result.Error)
	}

	t.ContainerId = result.ContainerId
	fmt.Printf("task %s is running in container %s\n", t.ID, t.ContainerId)
	fmt.Println("Sleepy time")
	time.Sleep(time.Second * 5)

	fmt.Printf("stopping task %s\n", t.ID)
	t.State = task.Completed
	w.AddTask(t)
	result = w.RunTask()
	if result.Error != nil {
		panic(result.Error)
	}
}

func createContainer() (*task.Docker, *task.DockerResult) {
	c := task.Config{
		Name:  "test-container-1",
		Image: "postgres:13",
		Env:   []string{"POSTGRES_USER=cube", "POSTGRES_PASSWORD=secret"},
	}
	dc, _ := client.NewClientWithOpts(client.FromEnv)
	d := task.Docker{
		Client: dc,
		Config: c,
	}
	runResult := d.Run()
	if runResult.Error != nil {
		fmt.Printf("Error running container: %v\n", runResult.Error)
		return nil, nil
	}

	fmt.Printf("Container %s is running with config %v\n", runResult.ContainerId, c)
	return &d, &runResult
}

func stopContainer(d *task.Docker, id string) *task.DockerResult {
	result := d.Stop(id)
	if result.Error != nil {
		fmt.Printf("Error stopping container: %v\n", result.Error)
		return nil
	}

	fmt.Printf("Container %s is stopped and removed\n", id)
	return &result
}

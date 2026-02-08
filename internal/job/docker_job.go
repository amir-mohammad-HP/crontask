// internal/job/docker_job.go
package job

import (
	"fmt"
	"sync"
	"time"

	"github.com/amir-mohammad-HP/crontask/pkg/docker"
	"github.com/robfig/cron/v3"
)

type DockerJob struct {
	id          string
	containerID string
	name        string
	cronExpr    string
	task        string
	monitor     *docker.DockerMonitor
	cronEntryID cron.EntryID
	lastRun     *time.Time
	nextRun     time.Time
}

func NewDockerJob(
	containerID string,
	name string,
	cronExpr string,
	task string,
	monitor *docker.DockerMonitor,
) *DockerJob {
	return &DockerJob{
		id:          fmt.Sprintf("%s-%s", containerID[:12], name),
		containerID: containerID,
		name:        name,
		cronExpr:    cronExpr,
		task:        task,
		monitor:     monitor,
	}
}

func (dj *DockerJob) Execute() error {
	dj.lastRun = &time.Time{}
	*dj.lastRun = time.Now()

	output, err := dj.monitor.ExecuteTask(dj.containerID, dj.task)
	if err != nil {
		return fmt.Errorf("failed to execute task in container %s: %w",
			dj.containerID[:12], err)
	}

	// Log output for debugging
	if len(output) > 0 {
		// You might want to log this or store it somewhere
		_ = output
	}

	return nil
}

func (dj *DockerJob) Name() string {
	return fmt.Sprintf("docker-%s", dj.id)
}

func (dj *DockerJob) Schedule() string {
	return dj.cronExpr
}

func (dj *DockerJob) GetContainerID() string {
	return dj.containerID
}

func (dj *DockerJob) SetCronEntryID(id cron.EntryID) {
	dj.cronEntryID = id
}

func (dj *DockerJob) GetCronEntryID() cron.EntryID {
	return dj.cronEntryID
}

func (dj *DockerJob) UpdateNextRun(schedule cron.Schedule) {
	dj.nextRun = schedule.Next(time.Now())
}

func (dj *DockerJob) GetLastRun() *time.Time {
	return dj.lastRun
}

func (dj *DockerJob) GetNextRun() time.Time {
	return dj.nextRun
}

// JobRegistry manages Docker jobs
type JobRegistry struct {
	jobs    map[string]*DockerJob
	mu      sync.RWMutex
	monitor *docker.DockerMonitor
}

func NewJobRegistry(monitor *docker.DockerMonitor) *JobRegistry {
	return &JobRegistry{
		jobs:    make(map[string]*DockerJob),
		monitor: monitor,
	}
}

func (jr *JobRegistry) AddJob(job *DockerJob) bool {
	jr.mu.Lock()
	defer jr.mu.Unlock()

	if _, exists := jr.jobs[job.id]; exists {
		return false
	}

	jr.jobs[job.id] = job
	return true
}

func (jr *JobRegistry) RemoveJob(jobID string) bool {
	jr.mu.Lock()
	defer jr.mu.Unlock()

	if _, exists := jr.jobs[jobID]; exists {
		delete(jr.jobs, jobID)
		return true
	}

	return false
}

func (jr *JobRegistry) RemoveJobsByContainer(containerID string) []string {
	jr.mu.Lock()
	defer jr.mu.Unlock()

	var removed []string
	for id, job := range jr.jobs {
		if job.containerID == containerID {
			delete(jr.jobs, id)
			removed = append(removed, id)
		}
	}

	return removed
}

func (jr *JobRegistry) GetJob(jobID string) (*DockerJob, bool) {
	jr.mu.RLock()
	defer jr.mu.RUnlock()

	job, exists := jr.jobs[jobID]
	return job, exists
}

func (jr *JobRegistry) GetAllJobs() []*DockerJob {
	jr.mu.RLock()
	defer jr.mu.RUnlock()

	jobs := make([]*DockerJob, 0, len(jr.jobs))
	for _, job := range jr.jobs {
		jobs = append(jobs, job)
	}

	return jobs
}

func (jr *JobRegistry) Count() int {
	jr.mu.RLock()
	defer jr.mu.RUnlock()
	return len(jr.jobs)
}

package homescript

import "sync"

// The manager is used to manage Homescripts and running processes

type HMSInitiator string

const (
	ApiInitiator        HMSInitiator = "api"
	AutomationInitiator HMSInitiator = "automation"
	HmsExecInitiator    HMSInitiator = "exec"
)

// Is used to manage the Homescript execution jobs
type Manager struct {
	Jobs           []Job
	LastInsertedId uint
	Lock           sync.RWMutex
}

// Represents a single HMS job
type Job struct {
	Id        uint
	Initiator HMSInitiator
	Executor  *Executor
	SigTerm   *chan int
}

// Appends a job to the manager
func (m *Manager) pushJob(initiator HMSInitiator, executor *Executor, sigTerm *chan int) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	// Increment the last inserted id by 1
	m.LastInsertedId++

	// Append the job to the slice
	m.Jobs = append(m.Jobs, Job{
		Id:        m.LastInsertedId,
		Initiator: initiator,
		Executor:  executor,
		SigTerm:   sigTerm,
	})
}

// Removes a job from the manager
// The returned boolean indicates whether the requested job has been removed or not
func (m *Manager) removeJob(jobId uint) bool {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	jobsTemp := make([]Job, 0)
	deleted := false

	for _, job := range m.Jobs {
		if job.Id != jobId {
			jobsTemp = append(jobsTemp, job)
		} else {
			deleted = true
		}
	}
	m.Jobs = jobsTemp
	return deleted
}

// Kills a running job using the HMS SigTerm
func (m *Manager) KillJob(jobId uint) bool {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	return false
}

package dispatcher

import (
	"sync"

	dispatcherTypes "github.com/smarthome-go/smarthome/core/homescript/dispatcher/types"
)

//
// Queue.
//

type PendingQueue struct {
	internal []dispatcherTypes.RegisterInfo
	lock     sync.Mutex
}

func NewQueue() PendingQueue {
	return PendingQueue{
		internal: make([]dispatcherTypes.RegisterInfo, 0),
		lock:     sync.Mutex{},
	}
}

func (q *PendingQueue) Enqueue(add dispatcherTypes.RegisterInfo) {
	q.lock.Lock()
	q.internal = append(q.internal, add)
	q.lock.Unlock()
}

func (q *PendingQueue) Dequeue() *dispatcherTypes.RegisterInfo {
	first := q.First()
	if first == nil {
		return nil
	}

	q.lock.Lock()
	q.internal = q.internal[1:]
	q.lock.Unlock()

	return first
}

func (q *PendingQueue) First() *dispatcherTypes.RegisterInfo {
	q.lock.Lock()
	defer q.lock.Unlock()

	if len(q.internal) == 0 {
		return nil
	}

	first := q.internal[0]
	return &first
}

func (q *PendingQueue) IsEmpty() bool {
	q.lock.Lock()
	defer q.lock.Unlock()

	return len(q.internal) == 0
}

func (q *PendingQueue) Len() int {
	q.lock.Lock()
	defer q.lock.Unlock()

	return len(q.internal)
}

//
// Dispatcher queue code.
//

func (i *InstanceT) RegisterPending() error {
	logger.Debug("Trying to register pending registrations...")

	var generalErr error

	pendingCount := i.PendingRegistrations.Len()
	for attempt := 0; attempt < pendingCount; attempt++ {
		current := i.PendingRegistrations.Dequeue()
		if current == nil {
			break
		}

		id, err := i.registerInternal(*current)
		if err != nil {
			if generalErr == nil {
				generalErr = err
			}
			i.PendingRegistrations.Enqueue(*current)
			continue
		}

		logger.Tracef("Successfully registered pending registration (new id: %d)\n", id)
	}

	return generalErr
}

func (i *InstanceT) Register(
	info dispatcherTypes.RegisterInfo,
	tolerance dispatcherTypes.Tolerance,
) (dispatcherTypes.RegistrationID, error) {
	id, err := i.registerInternal(info)

	if err == nil {
		return id, nil
	}

	if tolerance == dispatcherTypes.NoTolerance {
		return 0, err
	}

	// Retry this registration if feasible.
	i.PendingRegistrations.Enqueue(info)

	return 0, err
}

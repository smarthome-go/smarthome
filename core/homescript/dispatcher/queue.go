package dispatcher

import dispatcherTypes "github.com/smarthome-go/smarthome/core/homescript/dispatcher/types"

//
// Queue.
//

type PendingQueue struct {
	internal []dispatcherTypes.RegisterInfo
}

func (q *PendingQueue) Enqueue(add dispatcherTypes.RegisterInfo) {
	q.internal = append(q.internal, add)
}

func (q *PendingQueue) Dequeue() *dispatcherTypes.RegisterInfo {
	first := q.First()
	if first == nil {
		return nil
	}

	q.internal = q.internal[1:]

	return first
}

func (q *PendingQueue) First() *dispatcherTypes.RegisterInfo {
	if len(q.internal) == 0 {
		return nil
	}

	first := q.internal[0]
	return &first
}

func (q *PendingQueue) IsEmpty() bool {
	return len(q.internal) == 0
}

//
// Dispatcher queue code.
//

func (i *InstanceT) RegisterPending() error {
	logger.Debug("Trying to register pending registrations...")

	var generalErr error

	for !i.PendingRegistrations.IsEmpty() {
		current := i.PendingRegistrations.First()

		id, err := i.registerInternal(*current)
		if err != nil {
			if generalErr == nil {
				generalErr = err
			}
			continue
		}

		logger.Tracef("Successfully registered pending registration (new id: %d)\n", id)

		i.PendingRegistrations.Dequeue()
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

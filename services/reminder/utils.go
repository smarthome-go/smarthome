package reminder

import "github.com/MikMuellerDev/smarthome/core/database"

type Reminder struct {
	Id                uint                          `json:"id"`
	Name              string                        `json:"name"`
	Description       string                        `json:"description"`
	Priority          database.NotificationPriority `json:"priority"`
	CreatedDate       int64                         `json:"createdDate"` // Dates are represented as unix millis for making access in frontend easier
	DueDate           int64                         `json:"dueDate"`
	Owner             string                        `json:"owner"`
	UserWasNotified   bool                          `json:"userWasNotified"`
	UserWasNotifiedAt int64                         `json:"userWasNotifiedAt"`
}

// Returns a users reminders but transforms the underlying date to a unix timestamp
func GetUserReminders(username string) ([]Reminder, error) {
	reminders, err := database.GetUserReminders(username)
	if err != nil {
		return nil, err
	}
	remindersTemp := make([]Reminder, 0)
	for _, r := range reminders {
		remindersTemp = append(remindersTemp, Reminder{
			Id:                r.Id,
			Name:              r.Name,
			Description:       r.Description,
			Priority:          r.Priority,
			CreatedDate:       r.CreatedDate.UnixMilli(),
			DueDate:           r.DueDate.UnixMilli(),
			Owner:             r.Owner,
			UserWasNotified:   r.UserWasNotified,
			UserWasNotifiedAt: r.UserWasNotifiedAt.UnixMilli(),
		})
	}
	return remindersTemp, nil
}

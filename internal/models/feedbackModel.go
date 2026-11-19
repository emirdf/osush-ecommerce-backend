package models

import "time"

type Feedback struct {
	ID        int
	ClientID  *int
	Message   string
	CreatedAt time.Time

	// Filled only when the feedback is linked to an existing client.
	ClientUserName    *string
	ClientSurname     *string
	ClientPhoneNumber *string
}

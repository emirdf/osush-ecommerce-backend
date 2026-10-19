package models

type ClientNotification struct {
	ID             int
	DeviceTokenID  int
	NotificationID int
	IsRead         bool
}

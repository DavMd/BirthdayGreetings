package models

type Subscription struct {
	ID               int64 `json:"id" db:"id"`
	SubscriberID     int64 `json:"subscriber_id" db:"subscriber_id"`
	SubscribedUserID int64 `json:"subscribed_user_id" db:"subscribed_user_id"`
}

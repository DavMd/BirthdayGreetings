package subscription

import (
	"BirthdayGreetings/internal/db"
	"BirthdayGreetings/internal/models"
)

type SubscriptionService struct{}

func NewSubscriptionService() *SubscriptionService {
	return &SubscriptionService{}
}

func (s *SubscriptionService) SubscribeUser(subscriberID, subscribedUserID int64) error {

	sub := &models.Subscription{
		SubscriberID:     subscriberID,
		SubscribedUserID: subscribedUserID,
	}
	return db.CreateSubscription(sub)
}

func (s *SubscriptionService) UnsubscribeUser(subscriberID, subscribedUserID int64) error {
	return db.DeleteSubscription(subscriberID, subscribedUserID)
}

func (s *SubscriptionService) GetSubscriptions(userID int64) ([]models.UserBirthLayout, error) {
	subscriptions, err := db.GetSubscribers(userID)
	return subscriptions, err
}

func (s *SubscriptionService) IsSubscribed(subscriberID, subscribedUserID int64) (bool, error) {
	result, err := db.IsSubscribed(subscriberID, subscribedUserID)
	if err != nil {
		return false, err
	}
	return result, nil
}

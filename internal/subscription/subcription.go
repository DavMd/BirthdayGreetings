package subscription

import (
	"BirthdayGreetings/internal/db"
	"BirthdayGreetings/internal/models"
)

type SubscriptionService struct{}

func NewSubscriptionService() *SubscriptionService {
	return &SubscriptionService{}
}

func (s *SubscriptionService) CreateSubscription(subscriberID, subscribedUserID int64) (*models.Subscription, error) {
	sub := &models.Subscription{
		SubscriberID:     subscriberID,
		SubscribedUserID: subscribedUserID,
	}
	if err := db.CreateSubscription(sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *SubscriptionService) DeleteSubscription(subscriberID, subscribedUserID int64) error {
	if err := db.DeleteSubscription(subscriberID, subscribedUserID); err != nil {
		return err
	}
	return nil
}

func (s *SubscriptionService) GetSubscriptions(userID int64) ([]models.Subscription, error) {
	subscriptions, err := db.GetSubscriptions(userID)
	if err != nil {
		return nil, err
	}
	return subscriptions, nil
}

func (s *SubscriptionService) IsSubscribed(subscriberID, subscribedUserID int64) (bool, error) {
	result, err := db.IsSubscribed(subscriberID, subscribedUserID)
	if err != nil {
		return false, err
	}
	return result, nil
}

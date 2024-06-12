package telegram

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

type Client struct {
	client *telegram.Client
}

func NewClient(appID int, appHash, phoneNumber, password string) (*Client, error) {
	client := telegram.NewClient(appID, appHash, telegram.Options{})

	err := client.Run(context.Background(), func(ctx context.Context) error {
		codePrompt := func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
			fmt.Print("Введите код для доступа в аккаунт: ")
			code, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(code), nil
		}

		if err := auth.NewFlow(
			auth.Constant(phoneNumber, password, auth.CodeAuthenticatorFunc(codePrompt)),
			auth.SendCodeOptions{},
		).Run(ctx, client.Auth()); err != nil {
			panic(err)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	return &Client{client: client}, nil
}

func (c *Client) CreateChannel(ctx context.Context, title, about string) (*tg.Channel, error) {
	api := c.client.API()
	updates, err := api.ChannelsCreateChannel(ctx, &tg.ChannelsCreateChannelRequest{
		Broadcast: true,
		Megagroup: false,
		Title:     title,
		About:     about,
	})
	if err != nil {
		return nil, err
	}

	return getChannelFromUpdates(updates)
}

func (c *Client) AddUsersToChannel(ctx context.Context, channel *tg.Channel, userIDs []int64) error {
	api := c.client.API()
	inputUsers := make([]tg.InputUserClass, len(userIDs))

	for i, id := range userIDs {
		user, err := api.UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUser{
			UserID:     id,
			AccessHash: 0,
		}})
		if err != nil {
			return err
		}
		if len(user) == 0 {
			return fmt.Errorf("пользователь не найден: %d", id)
		}

		userClass := user[0].(*tg.User)
		inputUsers[i] = &tg.InputUser{
			UserID:     userClass.ID,
			AccessHash: userClass.AccessHash,
		}
	}

	_, err := api.ChannelsInviteToChannel(ctx, &tg.ChannelsInviteToChannelRequest{
		Channel: &tg.InputChannel{
			ChannelID:  channel.ID,
			AccessHash: channel.AccessHash,
		},
		Users: inputUsers,
	})
	return err
}

func (c *Client) AddBotToChannel(ctx context.Context, channel *tg.Channel, botID int64) error {
	api := c.client.API()
	_, err := api.ChannelsEditAdmin(ctx, &tg.ChannelsEditAdminRequest{
		Channel: &tg.InputChannel{ChannelID: channel.ID, AccessHash: channel.AccessHash},
		UserID:  &tg.InputUser{UserID: botID},
		Rank:    "admin",
		AdminRights: tg.ChatAdminRights{
			ChangeInfo:     true,
			DeleteMessages: true,
			InviteUsers:    true,
			PinMessages:    true,
		}})
	return err
}

func getChannelFromUpdates(updates tg.UpdatesClass) (*tg.Channel, error) {
	switch u := updates.(type) {
	case *tg.Updates:
		for _, chat := range u.Chats {
			if channel, ok := chat.(*tg.Channel); ok {
				return channel, nil
			}
		}
	case *tg.UpdatesCombined:
		for _, chat := range u.Chats {
			if channel, ok := chat.(*tg.Channel); ok {
				return channel, nil
			}
		}
	}
	return nil, fmt.Errorf("канал не найден")
}

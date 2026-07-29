package twitch

import (
	"context"
	"net/url"
	"strings"
)

type User struct {
	ID              string `json:"id"`
	Login           string `json:"login"`
	DisplayName     string `json:"display_name"`
	Description     string `json:"description"`
	ProfileImageURL string `json:"profile_image_url"`
	OfflineImageURL string `json:"offline_image_url"`
}

func (u User) Name() string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	return u.Login
}

func (c *Client) GetUsersByLogin(ctx context.Context, logins []string) ([]User, error) {
	normalized := make([]string, 0, len(logins))
	for _, l := range logins {
		normalized = append(normalized, strings.ToLower(strings.TrimSpace(l)))
	}
	return c.getUsers(ctx, "login", normalized)
}

func (c *Client) GetUsersByID(ctx context.Context, ids []string) ([]User, error) {
	return c.getUsers(ctx, "id", ids)
}

func (c *Client) GetUserByLogin(ctx context.Context, login string) (User, bool, error) {
	users, err := c.GetUsersByLogin(ctx, []string{login})
	if err != nil || len(users) == 0 {
		return User{}, false, err
	}
	return users[0], true, nil
}

func (c *Client) GetUserByID(ctx context.Context, id string) (User, bool, error) {
	users, err := c.GetUsersByID(ctx, []string{id})
	if err != nil || len(users) == 0 {
		return User{}, false, err
	}
	return users[0], true, nil
}

func (c *Client) getUsers(ctx context.Context, param string, values []string) ([]User, error) {
	values = dedupe(values)
	if len(values) == 0 {
		return nil, nil
	}

	out := make([]User, 0, len(values))
	for _, batch := range chunk(values, MaxIDsPerRequest) {
		query := url.Values{}
		for _, v := range batch {
			query.Add(param, v)
		}

		var resp response[User]
		if err := c.get(ctx, "/users", query, &resp); err != nil {
			return nil, err
		}
		out = append(out, resp.Data...)
	}
	return out, nil
}

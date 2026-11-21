package yamusic

import (
	"context"
	"encoding/json"
	"net/url"
)

// Init loads /account/status and stores the account id for later calls
// that default to the current user (likes, playlists). A guest token yields
// a status whose account id stays empty.
func (c *Client) Init(ctx context.Context) (*Status, error) {
	st, err := c.AccountStatus(ctx)
	if err != nil {
		return nil, err
	}
	if st != nil {
		c.setUID(st.Account.UID.String())
	}
	return st, nil
}

// AccountStatus returns the current account, permissions, and subscription.
func (c *Client) AccountStatus(ctx context.Context) (*Status, error) {
	raw, err := c.get(ctx, "/account/status", nil)
	if err != nil {
		return nil, err
	}
	st, err := decode[Status](raw)
	if err != nil {
		return nil, err
	}
	return &st, nil
}

// AccountSettings returns the current user's playback settings.
func (c *Client) AccountSettings(ctx context.Context) (*UserSettings, error) {
	raw, err := c.get(ctx, "/account/settings", nil)
	if err != nil {
		return nil, err
	}
	s, err := decode[UserSettings](raw)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// SetAccountSettings writes one settings field. name is the camelCase field
// the API expects (for example "shuffleEnabled"). value is sent as a string,
// which is how the settings endpoint reads booleans and numbers.
func (c *Client) SetAccountSettings(ctx context.Context, name, value string) (*UserSettings, error) {
	if name == "" {
		return nil, &Error{Message: "settings field name is empty"}
	}
	// The body is one form field whose name is the setting, not a nested object.
	form := url.Values{}
	form.Set(name, value)
	raw, err := c.postForm(ctx, "/account/settings", form)
	if err != nil {
		return nil, err
	}
	s, err := decode[UserSettings](raw)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Experiments returns the account's experiment groups as a JSON object.
// The keys are experiment names and the values are the assigned group.
// The object is not modelled further: the server adds experiments freely.
func (c *Client) Experiments(ctx context.Context) (map[string]json.RawMessage, error) {
	raw, err := c.get(ctx, "/account/experiments", nil)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return map[string]json.RawMessage{}, nil
	}
	var out map[string]json.RawMessage
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, &Error{Message: "decode experiments: " + err.Error()}
	}
	if out == nil {
		out = map[string]json.RawMessage{}
	}
	return out, nil
}

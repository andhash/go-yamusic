package yamusic

import (
	"context"
	"net/url"
)

// postIDs posts a comma-separated id list. extra is copied into the form
// before the id field, so a caller can add with-positions without replacing ids.
func postIDs[T any](c *Client, ctx context.Context, path, field string, ids []string, extra url.Values) ([]T, error) {
	joined, err := joinIDs(ids)
	if err != nil {
		return nil, err
	}
	form := url.Values{}
	for k, vs := range extra {
		for _, v := range vs {
			form.Add(k, v)
		}
	}
	form.Set(field, joined)
	raw, err := c.postForm(ctx, path, form)
	if err != nil {
		return nil, err
	}
	return decodeList[T](raw)
}

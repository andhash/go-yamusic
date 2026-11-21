package yamusic

import (
	"context"
	"net/url"
	"strings"
)

// Genres returns the genre tree. SubGenres is empty for a leaf.
func (c *Client) Genres(ctx context.Context) ([]Genre, error) {
	raw, err := c.get(ctx, "/genres", nil)
	if err != nil {
		return nil, err
	}
	return decodeList[Genre](raw)
}

// Landing loads the home page. blocks limits which sections come back
// (for example "personalplaylists", "new-releases", "chart"). An empty list
// asks for the default set.
func (c *Client) Landing(ctx context.Context, blocks ...string) (*Landing, error) {
	q := url.Values{}
	if len(blocks) > 0 {
		q.Set("blocks", strings.Join(blocks, ","))
	}
	raw, err := c.get(ctx, "/landing3", q)
	if err != nil {
		return nil, err
	}
	out, err := decode[Landing](raw)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Chart returns the chart playlist. option is a regional suffix such as
// "russia" or "world". An empty option is the default chart.
func (c *Client) Chart(ctx context.Context, option string) (*Chart, error) {
	path := "/landing3/chart"
	if option != "" {
		path += "/" + url.PathEscape(option)
	}
	raw, err := c.get(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	out, err := decode[Chart](raw)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// NewReleases returns the landing list of new albums.
func (c *Client) NewReleases(ctx context.Context) ([]Album, error) {
	raw, err := c.get(ctx, "/landing3/new-releases", nil)
	if err != nil {
		return nil, err
	}
	// The block is either a bare list or {"newReleases":[...]}.
	if isJSONObject(raw) {
		out, err := decode[struct {
			NewReleases []Album `json:"newReleases"`
		}](raw)
		if err != nil {
			return nil, err
		}
		if out.NewReleases == nil {
			return []Album{}, nil
		}
		return out.NewReleases, nil
	}
	return decodeList[Album](raw)
}

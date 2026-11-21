package yamusic

import (
	"context"
	"net/url"
)

// Album returns one album without its track list.
func (c *Client) Album(ctx context.Context, id string) (*Album, error) {
	list, err := c.Albums(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, &Error{Status: 404, Name: "not-found", Message: "album not found"}
	}
	return &list[0], nil
}

// Albums loads albums by id. The body field is album-ids.
func (c *Client) Albums(ctx context.Context, ids ...string) ([]Album, error) {
	return postIDs[Album](c, ctx, "/albums", "album-ids", ids, nil)
}

// AlbumWithTracks returns the album and its discs. Volumes[i] is disc i.
func (c *Client) AlbumWithTracks(ctx context.Context, id string) (*Album, error) {
	if id == "" {
		return nil, &Error{Message: "album id is empty"}
	}
	raw, err := c.get(ctx, "/albums/"+url.PathEscape(id)+"/with-tracks", nil)
	if err != nil {
		return nil, err
	}
	out, err := decode[Album](raw)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

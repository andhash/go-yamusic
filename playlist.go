package yamusic

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
)

// Visibility is the playlist visibility value the API stores.
type Visibility string

const (
	// VisibilityPublic is a playlist anyone can open.
	VisibilityPublic Visibility = "public"
	// VisibilityPrivate is a playlist only the owner can open.
	VisibilityPrivate Visibility = "private"
)

// UserPlaylists returns the owner's playlists. An empty kinds list returns
// every playlist of that user. A non-empty list asks only for those kind numbers.
func (c *Client) UserPlaylists(ctx context.Context, userID string, kinds ...int) ([]Playlist, error) {
	userID, err := c.user(userID)
	if err != nil {
		return nil, err
	}
	if len(kinds) == 0 {
		raw, err := c.get(ctx, "/users/"+url.PathEscape(userID)+"/playlists/list", nil)
		if err != nil {
			return nil, err
		}
		return decodeList[Playlist](raw)
	}
	form := url.Values{}
	ids := make([]string, len(kinds))
	for i, k := range kinds {
		ids[i] = strconv.Itoa(k)
	}
	joined, err := joinIDs(ids)
	if err != nil {
		return nil, err
	}
	form.Set("kinds", joined)
	raw, err := c.postForm(ctx, "/users/"+url.PathEscape(userID)+"/playlists", form)
	if err != nil {
		return nil, err
	}
	return decodeList[Playlist](raw)
}

// Playlist returns one playlist of userID with the given kind, including its tracks.
func (c *Client) Playlist(ctx context.Context, userID string, kind int) (*Playlist, error) {
	userID, err := c.user(userID)
	if err != nil {
		return nil, err
	}
	path := "/users/" + url.PathEscape(userID) + "/playlists/" + strconv.Itoa(kind)
	raw, err := c.get(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	out, err := decode[Playlist](raw)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// CreatePlaylist creates a playlist and returns it. visibility empty means private.
func (c *Client) CreatePlaylist(ctx context.Context, userID, title string, visibility Visibility) (*Playlist, error) {
	userID, err := c.user(userID)
	if err != nil {
		return nil, err
	}
	if title == "" {
		return nil, &Error{Message: "playlist title is empty"}
	}
	if visibility == "" {
		visibility = VisibilityPrivate
	}
	form := url.Values{}
	form.Set("title", title)
	form.Set("visibility", string(visibility))
	raw, err := c.postForm(ctx, "/users/"+url.PathEscape(userID)+"/playlists/create", form)
	if err != nil {
		return nil, err
	}
	out, err := decode[Playlist](raw)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// RenamePlaylist sets the playlist title. The API expects the new title in the form field "value".
func (c *Client) RenamePlaylist(ctx context.Context, userID string, kind int, title string) (*Playlist, error) {
	return c.playlistValue(ctx, userID, kind, "name", title)
}

// SetPlaylistVisibility sets public or private.
func (c *Client) SetPlaylistVisibility(ctx context.Context, userID string, kind int, visibility Visibility) (*Playlist, error) {
	if visibility != VisibilityPublic && visibility != VisibilityPrivate {
		return nil, &Error{Message: "visibility must be public or private"}
	}
	return c.playlistValue(ctx, userID, kind, "visibility", string(visibility))
}

func (c *Client) playlistValue(ctx context.Context, userID string, kind int, action, value string) (*Playlist, error) {
	userID, err := c.user(userID)
	if err != nil {
		return nil, err
	}
	path := "/users/" + url.PathEscape(userID) + "/playlists/" + strconv.Itoa(kind) + "/" + action
	form := url.Values{}
	form.Set("value", value)
	raw, err := c.postForm(ctx, path, form)
	if err != nil {
		return nil, err
	}
	out, err := decode[Playlist](raw)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// DeletePlaylist removes the playlist.
func (c *Client) DeletePlaylist(ctx context.Context, userID string, kind int) error {
	userID, err := c.user(userID)
	if err != nil {
		return err
	}
	path := "/users/" + url.PathEscape(userID) + "/playlists/" + strconv.Itoa(kind) + "/delete"
	_, err = c.postForm(ctx, path, url.Values{})
	return err
}

// InsertTracks inserts tracks at position at. at 0 is the front of the playlist.
// revision is the playlist's current Revision; the server rejects a stale one.
// A TrackRef.AlbumID may be empty when the id already has the form "track:album".
func (c *Client) InsertTracks(ctx context.Context, userID string, kind, revision, at int, tracks []TrackRef) (*Playlist, error) {
	op := struct {
		Op     string     `json:"op"`
		At     int        `json:"at"`
		Tracks []TrackRef `json:"tracks"`
	}{Op: "insert", At: at, Tracks: normalizeRefs(tracks)}
	return c.changePlaylist(ctx, userID, kind, revision, []any{op})
}

// DeleteTrackRange removes playlist rows [from, to). revision is the current Revision.
func (c *Client) DeleteTrackRange(ctx context.Context, userID string, kind, revision, from, to int) (*Playlist, error) {
	if to < from {
		return nil, &Error{Message: "track range is reversed"}
	}
	op := struct {
		Op   string `json:"op"`
		From int    `json:"from"`
		To   int    `json:"to"`
	}{Op: "delete", From: from, To: to}
	return c.changePlaylist(ctx, userID, kind, revision, []any{op})
}

func (c *Client) changePlaylist(ctx context.Context, userID string, kind, revision int, diff any) (*Playlist, error) {
	userID, err := c.user(userID)
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(diff)
	if err != nil {
		return nil, err
	}
	form := url.Values{}
	form.Set("kind", strconv.Itoa(kind))
	form.Set("revision", strconv.Itoa(revision))
	form.Set("diff", string(encoded))
	path := "/users/" + url.PathEscape(userID) + "/playlists/" + strconv.Itoa(kind) + "/change"
	raw, err := c.postForm(ctx, path, form)
	if err != nil {
		return nil, err
	}
	out, err := decode[Playlist](raw)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// normalizeRefs copies refs and fills AlbumID from a "track:album" id when the album is missing.
func normalizeRefs(tracks []TrackRef) []TrackRef {
	out := make([]TrackRef, len(tracks))
	for i, t := range tracks {
		if t.AlbumID == "" {
			id, album := SplitTrackID(t.ID)
			if album != "" {
				t.ID = id
				t.AlbumID = album
			}
		}
		out[i] = t
	}
	return out
}

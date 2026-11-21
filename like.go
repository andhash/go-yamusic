package yamusic

import (
	"context"
	"net/url"
	"strconv"
)

// Like marks ids as liked for the user. objectType is track, album, artist, or playlist.
// A playlist id is "ownerUid:kind". An empty userID uses the id stored by Init.
// Liking a track also clears a dislike on that track.
func (c *Client) Like(ctx context.Context, userID, objectType string, ids ...string) error {
	return c.likeAction(ctx, "likes", userID, objectType, false, ids)
}

// Unlike removes the like mark.
func (c *Client) Unlike(ctx context.Context, userID, objectType string, ids ...string) error {
	return c.likeAction(ctx, "likes", userID, objectType, true, ids)
}

// Dislike marks ids as "do not recommend". objectType is track or artist.
// Disliking a track also clears a like on that track.
func (c *Client) Dislike(ctx context.Context, userID, objectType string, ids ...string) error {
	return c.likeAction(ctx, "dislikes", userID, objectType, false, ids)
}

// Undislike removes the dislike mark.
func (c *Client) Undislike(ctx context.Context, userID, objectType string, ids ...string) error {
	return c.likeAction(ctx, "dislikes", userID, objectType, true, ids)
}

func (c *Client) likeAction(ctx context.Context, bucket, userID, objectType string, remove bool, ids []string) error {
	switch objectType {
	case "track", "album", "artist", "playlist":
	default:
		return &Error{Message: "object type must be track, album, artist, or playlist"}
	}
	if bucket == "dislikes" && objectType != "track" && objectType != "artist" {
		return &Error{Message: "dislike accepts track or artist"}
	}
	userID, err := c.user(userID)
	if err != nil {
		return err
	}
	joined, err := joinIDs(ids)
	if err != nil {
		return err
	}
	action := "add-multiple"
	if remove {
		action = "remove"
	}
	path := "/users/" + url.PathEscape(userID) + "/" + bucket + "/" + objectType + "s/" + action
	form := url.Values{}
	form.Set(objectType+"-ids", joined)
	_, err = c.postForm(ctx, path, form)
	return err
}

// LikedTracks returns the user's liked track ids. revision 0 returns the full list.
// A non-zero revision asks the server to skip the body when nothing changed since that revision.
func (c *Client) LikedTracks(ctx context.Context, userID string, revision int) (*Library, error) {
	return c.library(ctx, "likes", userID, revision, "if-modified-since-revision")
}

// DislikedTracks returns the user's disliked track ids.
func (c *Client) DislikedTracks(ctx context.Context, userID string, revision int) (*Library, error) {
	return c.library(ctx, "dislikes", userID, revision, "if_modified_since_revision")
}

func (c *Client) library(ctx context.Context, bucket, userID string, revision int, revisionParam string) (*Library, error) {
	userID, err := c.user(userID)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set(revisionParam, strconv.Itoa(revision))
	raw, err := c.get(ctx, "/users/"+url.PathEscape(userID)+"/"+bucket+"/tracks", q)
	if err != nil {
		return nil, err
	}
	var wrap struct {
		Library *Library `json:"library"`
	}
	wrap, err = decode[struct {
		Library *Library `json:"library"`
	}](raw)
	if err != nil {
		return nil, err
	}
	if wrap.Library == nil {
		return &Library{}, nil
	}
	return wrap.Library, nil
}

// LikedAlbums returns liked albums. rich false asks for the short form.
func (c *Client) LikedAlbums(ctx context.Context, userID string, rich bool) ([]Like, error) {
	q := url.Values{}
	q.Set("rich", strconv.FormatBool(rich))
	return c.likes(ctx, userID, "albums", q)
}

// LikedArtists returns liked artists.
func (c *Client) LikedArtists(ctx context.Context, userID string) ([]Like, error) {
	q := url.Values{}
	q.Set("with-timestamps", "true")
	return c.likes(ctx, userID, "artists", q)
}

// LikedPlaylists returns liked playlists.
func (c *Client) LikedPlaylists(ctx context.Context, userID string) ([]Like, error) {
	return c.likes(ctx, userID, "playlists", nil)
}

func (c *Client) likes(ctx context.Context, userID, plural string, query url.Values) ([]Like, error) {
	userID, err := c.user(userID)
	if err != nil {
		return nil, err
	}
	raw, err := c.get(ctx, "/users/"+url.PathEscape(userID)+"/likes/"+plural, query)
	if err != nil {
		return nil, err
	}
	return decodeList[Like](raw)
}

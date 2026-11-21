package yamusic

import (
	"context"
	"net/url"
)

// Track returns one track. id is a track id or "trackId:albumId".
// A response with an empty list is a not-found error.
func (c *Client) Track(ctx context.Context, id string) (*Track, error) {
	list, err := c.Tracks(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, &Error{Status: 404, Name: "not-found", Message: "track not found"}
	}
	return &list[0], nil
}

// Tracks loads tracks by id. The POST body is track-ids, comma-separated,
// plus with-positions=true so album positions stay attached.
func (c *Client) Tracks(ctx context.Context, ids ...string) ([]Track, error) {
	return postIDs[Track](c, ctx, "/tracks", "track-ids", ids, url.Values{"with-positions": {"true"}})
}

// TrackDownloadInfo lists the file variants the API offers for trackID.
// Each item's DownloadInfoURL is the address the server returned.
// This method does not sign that address and does not download the bytes.
func (c *Client) TrackDownloadInfo(ctx context.Context, trackID string) ([]DownloadInfo, error) {
	if trackID == "" {
		return nil, &Error{Message: "track id is empty"}
	}
	raw, err := c.get(ctx, "/tracks/"+url.PathEscape(trackID)+"/download-info", nil)
	if err != nil {
		return nil, err
	}
	return decodeList[DownloadInfo](raw)
}

// SimilarTracks returns tracks the catalogue marks as similar to trackID.
type SimilarTracks struct {
	Track   *Track  `json:"track"`
	Similar []Track `json:"similarTracks"`
}

// Similar returns the similar-track block for trackID.
func (c *Client) Similar(ctx context.Context, trackID string) (*SimilarTracks, error) {
	if trackID == "" {
		return nil, &Error{Message: "track id is empty"}
	}
	raw, err := c.get(ctx, "/tracks/"+url.PathEscape(trackID)+"/similar", nil)
	if err != nil {
		return nil, err
	}
	out, err := decode[SimilarTracks](raw)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Supplement is the extra block from /tracks/{id}/supplement (description
// and credits). Lyrics in this block are stale; the current lyrics endpoint
// requires a request signature this package does not build.
type Supplement struct {
	ID          ID     `json:"id"`
	Lyrics      *Lyric `json:"lyrics"`
	Description string `json:"description"`
}

// Lyric is a text blob inside a supplement. FullText is the plain text when
// the server still sends it there.
type Lyric struct {
	ID        ID     `json:"id"`
	Lyrics    string `json:"lyrics"`
	FullText  string `json:"fullLyrics"`
	HasRights bool   `json:"hasRights"`
}

// Supplement returns the track supplement.
func (c *Client) Supplement(ctx context.Context, trackID string) (*Supplement, error) {
	if trackID == "" {
		return nil, &Error{Message: "track id is empty"}
	}
	raw, err := c.get(ctx, "/tracks/"+url.PathEscape(trackID)+"/supplement", nil)
	if err != nil {
		return nil, err
	}
	out, err := decode[Supplement](raw)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

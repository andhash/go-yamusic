package yamusic

import (
	"context"
	"net/url"
	"strconv"
)

// SearchType selects which catalogue objects /search returns.
type SearchType string

const (
	// SearchAll searches tracks, artists, albums, and playlists.
	// Podcasts are not included; use SearchPodcast or SearchPodcastEpisode.
	SearchAll            SearchType = "all"
	SearchTrack          SearchType = "track"
	SearchArtist         SearchType = "artist"
	SearchAlbum          SearchType = "album"
	SearchPlaylist       SearchType = "playlist"
	SearchUser           SearchType = "user"
	SearchPodcast        SearchType = "podcast"
	SearchPodcastEpisode SearchType = "podcast_episode"
)

// SearchQuery is one /search call. Page is zero-based. Zero Type means SearchAll.
// Nocorrect false lets the server fix a misspelled query.
type SearchQuery struct {
	Text           string
	Type           SearchType
	Page           int
	Nocorrect      bool
	PlaylistInBest bool
}

// Search runs a catalogue search.
func (c *Client) Search(ctx context.Context, q SearchQuery) (*Search, error) {
	if q.Text == "" {
		return nil, &Error{Message: "search text is empty"}
	}
	if q.Type == "" {
		q.Type = SearchAll
	}
	query := url.Values{}
	query.Set("text", q.Text)
	query.Set("type", string(q.Type))
	query.Set("page", strconv.Itoa(q.Page))
	query.Set("nocorrect", strconv.FormatBool(q.Nocorrect))
	query.Set("playlist-in-best", strconv.FormatBool(q.PlaylistInBest))
	raw, err := c.get(ctx, "/search", query)
	if err != nil {
		return nil, err
	}
	out, err := decode[Search](raw)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Suggest returns completion strings for a partial query.
func (c *Client) Suggest(ctx context.Context, part string) (*Suggestions, error) {
	if part == "" {
		return nil, &Error{Message: "suggest text is empty"}
	}
	raw, err := c.get(ctx, "/search/suggest", url.Values{"part": {part}})
	if err != nil {
		return nil, err
	}
	out, err := decode[Suggestions](raw)
	if err != nil {
		return nil, err
	}
	if out.Suggestions == nil {
		out.Suggestions = []string{}
	}
	return &out, nil
}

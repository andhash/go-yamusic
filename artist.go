package yamusic

import (
	"context"
	"net/url"
	"strconv"
)

// ArtistSort is the sort-by value for an artist's album lists.
type ArtistSort string

const (
	// SortByYear orders albums by release year.
	SortByYear ArtistSort = "year"
	// SortByRating orders albums by the catalogue rating.
	SortByRating ArtistSort = "rating"
)

// Artist returns one artist.
func (c *Client) Artist(ctx context.Context, id string) (*Artist, error) {
	list, err := c.Artists(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, &Error{Status: 404, Name: "not-found", Message: "artist not found"}
	}
	return &list[0], nil
}

// Artists loads artists by id. The body field is artist-ids.
func (c *Client) Artists(ctx context.Context, ids ...string) ([]Artist, error) {
	return postIDs[Artist](c, ctx, "/artists", "artist-ids", ids, nil)
}

// ArtistBrief is the combined card from /artists/{id}/brief-info.
type ArtistBrief struct {
	Artist         Artist   `json:"artist"`
	Albums         []Album  `json:"albums"`
	AlsoAlbums     []Album  `json:"alsoAlbums"`
	PopularTracks  []Track  `json:"popularTracks"`
	SimilarArtists []Artist `json:"similarArtists"`
}

// ArtistBriefInfo returns the artist card: popular tracks, albums, similar artists.
func (c *Client) ArtistBriefInfo(ctx context.Context, id string) (*ArtistBrief, error) {
	if id == "" {
		return nil, &Error{Message: "artist id is empty"}
	}
	raw, err := c.get(ctx, "/artists/"+url.PathEscape(id)+"/brief-info", nil)
	if err != nil {
		return nil, err
	}
	out, err := decode[ArtistBrief](raw)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ArtistTracks returns one page of the artist's own tracks.
// page is zero-based. pageSize 0 uses the server default.
func (c *Client) ArtistTracks(ctx context.Context, id string, page, pageSize int) (*ArtistTracks, error) {
	if id == "" {
		return nil, &Error{Message: "artist id is empty"}
	}
	q := url.Values{}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	if pageSize > 0 {
		q.Set("page-size", strconv.Itoa(pageSize))
	}
	raw, err := c.get(ctx, "/artists/"+url.PathEscape(id)+"/tracks", q)
	if err != nil {
		return nil, err
	}
	out, err := decode[ArtistTracks](raw)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ArtistAlbums returns one page of the artist's own albums.
func (c *Client) ArtistAlbums(ctx context.Context, id string, sort ArtistSort, page, pageSize int) (*ArtistAlbums, error) {
	return c.artistAlbumPage(ctx, id, "direct-albums", sort, page, pageSize)
}

// ArtistAlsoAlbums returns albums where the artist appears but is not the headliner.
func (c *Client) ArtistAlsoAlbums(ctx context.Context, id string, sort ArtistSort, page, pageSize int) (*ArtistAlbums, error) {
	return c.artistAlbumPage(ctx, id, "also-albums", sort, page, pageSize)
}

func (c *Client) artistAlbumPage(ctx context.Context, id, kind string, sort ArtistSort, page, pageSize int) (*ArtistAlbums, error) {
	if id == "" {
		return nil, &Error{Message: "artist id is empty"}
	}
	q := url.Values{}
	if sort != "" {
		q.Set("sort-by", string(sort))
	}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	if pageSize > 0 {
		q.Set("page-size", strconv.Itoa(pageSize))
	}
	raw, err := c.get(ctx, "/artists/"+url.PathEscape(id)+"/"+kind, q)
	if err != nil {
		return nil, err
	}
	out, err := decode[ArtistAlbums](raw)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// SimilarArtists returns artists the catalogue marks as similar.
func (c *Client) SimilarArtists(ctx context.Context, id string) ([]Artist, error) {
	if id == "" {
		return nil, &Error{Message: "artist id is empty"}
	}
	raw, err := c.get(ctx, "/artists/"+url.PathEscape(id)+"/similar", nil)
	if err != nil {
		return nil, err
	}
	// The payload is either a bare list or an object with an "artists" field.
	if isJSONObject(raw) {
		out, err := decode[struct {
			Artists []Artist `json:"artists"`
		}](raw)
		if err != nil {
			return nil, err
		}
		if out.Artists == nil {
			return []Artist{}, nil
		}
		return out.Artists, nil
	}
	return decodeList[Artist](raw)
}

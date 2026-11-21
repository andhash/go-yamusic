package yamusic

import (
	"encoding/json"
	"strings"
)

// Status is the account snapshot returned by /account/status and by Init.
type Status struct {
	Account      Account      `json:"account"`
	Permissions  Permissions  `json:"permissions"`
	Subscription Subscription `json:"subscription"`
}

// Account is the signed-in user. UID is empty for a guest response that
// has no account object.
type Account struct {
	UID              ID     `json:"uid"`
	Login            string `json:"login"`
	FullName         string `json:"fullName"`
	DisplayName      string `json:"displayName"`
	Region           int    `json:"region"`
	ServiceAvailable bool   `json:"serviceAvailable"`
	Now              string `json:"now"`
}

// Permissions lists what the account is allowed to play.
// Values contains entries such as "premium-plus" when the account has them.
type Permissions struct {
	Until  string   `json:"until"`
	Values []string `json:"values"`
}

// Has reports whether values contains name.
func (p Permissions) Has(name string) bool {
	for _, v := range p.Values {
		if v == name {
			return true
		}
	}
	return false
}

// Subscription is the paid-plan block. The nested product lists stay raw
// because the server adds fields without notice. NonAutoRenewable is the
// remainder object the API names nonAutoRenewableRemainder.
type Subscription struct {
	CanStartTrial      bool  `json:"canStartTrial"`
	Mcdonalds          bool  `json:"mcdonalds"`
	HadAnySubscription bool  `json:"hadAnySubscription"`
	AutoRenewable      []any `json:"autoRenewable"`
	NonAutoRenewable   any   `json:"nonAutoRenewableRemainder"`
}

// UserSettings is /account/settings. Only the stable fields are typed.
// The rest of the object is large and changes with experiments.
type UserSettings struct {
	UID                       ID     `json:"uid"`
	LastFMScrobblingEnabled   bool   `json:"lastFmScrobblingEnabled"`
	FacebookScrobblingEnabled bool   `json:"facebookScrobblingEnabled"`
	ShuffleEnabled            bool   `json:"shuffleEnabled"`
	VolumePercents            int    `json:"volumePercents"`
	UserMusicVisibility       string `json:"userMusicVisibility"`
}

// Track is one catalogue track. ID may be "trackId" or "trackId:albumId".
// Albums on a track do not include Volumes; those come from AlbumWithTracks.
type Track struct {
	ID                       ID       `json:"id"`
	RealID                   ID       `json:"realId"`
	Title                    string   `json:"title"`
	Version                  string   `json:"version"`
	Type                     string   `json:"type"`
	Available                bool     `json:"available"`
	AvailableForPremiumUsers bool     `json:"availableForPremiumUsers"`
	LyricsAvailable          bool     `json:"lyricsAvailable"`
	DurationMs               int64    `json:"durationMs"`
	CoverURI                 string   `json:"coverUri"`
	OgImage                  string   `json:"ogImage"`
	ContentWarning           string   `json:"contentWarning"`
	Artists                  []Artist `json:"artists"`
	Albums                   []Album  `json:"albums"`
}

// ArtistName joins the track's artist names with a comma.
func (t Track) ArtistName() string {
	names := make([]string, 0, len(t.Artists))
	for _, a := range t.Artists {
		if a.Name != "" {
			names = append(names, a.Name)
		}
	}
	return strings.Join(names, ", ")
}

// Artist is a performer. Cover is the template object; CoverURI on a track
// is the flat string used by CoverURL.
type Artist struct {
	ID        ID       `json:"id"`
	Name      string   `json:"name"`
	Various   bool     `json:"various"`
	Cover     *Cover   `json:"cover"`
	OgImage   string   `json:"ogImage"`
	Genres    []string `json:"genres"`
	Counts    *Counts  `json:"counts"`
	Available bool     `json:"available"`
}

// Counts is the artist's catalogue size.
type Counts struct {
	Tracks       int `json:"tracks"`
	DirectAlbums int `json:"directAlbums"`
	AlsoAlbums   int `json:"alsoAlbums"`
	AlsoTracks   int `json:"alsoTracks"`
}

// Cover is an image template. URI contains the literal "%%", which CoverURL replaces.
type Cover struct {
	URI    string `json:"uri"`
	Type   string `json:"type"`
	Prefix string `json:"prefix"`
	Custom bool   `json:"custom"`
}

// CoverURL builds a concrete image URL. size is a Yandex avatar size such
// as "200x200" or "1000x1000". uri is either a template with "%%" or an
// already absolute URL. An empty uri returns an empty string.
func CoverURL(uri, size string) string {
	uri = strings.TrimSpace(uri)
	if uri == "" {
		return ""
	}
	if size == "" {
		size = "200x200"
	}
	uri = strings.ReplaceAll(uri, "%%", size)
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		return uri
	}
	uri = strings.TrimPrefix(uri, "//")
	return "https://" + uri
}

// Album is a release. Volumes is set by AlbumWithTracks and is a list of
// discs, each disc a list of tracks. TrackCount is the catalogue count and
// stays set even when Volumes is empty.
type Album struct {
	ID          ID        `json:"id"`
	Title       string    `json:"title"`
	Version     string    `json:"version"`
	Type        string    `json:"type"`
	MetaType    string    `json:"metaType"`
	Year        int       `json:"year"`
	ReleaseDate string    `json:"releaseDate"`
	Genre       string    `json:"genre"`
	TrackCount  int       `json:"trackCount"`
	Available   bool      `json:"available"`
	CoverURI    string    `json:"coverUri"`
	OgImage     string    `json:"ogImage"`
	Artists     []Artist  `json:"artists"`
	Volumes     [][]Track `json:"volumes"`
	Labels      []Label   `json:"labels"`
}

// Label is a record label name attached to an album.
type Label struct {
	ID   ID     `json:"id"`
	Name string `json:"name"`
}

// Playlist is a user playlist or a chart playlist.
// Kind is the owner's local playlist number. Owner plus Kind identify it.
// Tracks is present when the request asked for the track list.
type Playlist struct {
	UID         ID              `json:"uid"`
	Kind        int             `json:"kind"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Revision    int             `json:"revision"`
	TrackCount  int             `json:"trackCount"`
	Visibility  string          `json:"visibility"`
	Available   bool            `json:"available"`
	Owner       *User           `json:"owner"`
	Cover       *Cover          `json:"cover"`
	Tracks      []PlaylistTrack `json:"tracks"`
}

// User is a playlist owner or another public profile.
type User struct {
	UID   ID     `json:"uid"`
	Login string `json:"login"`
	Name  string `json:"name"`
}

// PlaylistTrack is one row inside a playlist. Track is nil when the server
// returned only the id pair.
type PlaylistTrack struct {
	ID        ID     `json:"id"`
	Timestamp string `json:"timestamp"`
	Track     *Track `json:"track"`
}

// TrackRef is a track id plus the album id the playlist row points at.
type TrackRef struct {
	ID      string `json:"id"`
	AlbumID string `json:"albumId,omitempty"`
}

// Library is the "Мне нравится" track list. Revision is the value the next
// like or dislike call observes. Tracks holds id pairs, not full Track objects.
type Library struct {
	UID      ID         `json:"uid"`
	Revision int        `json:"revision"`
	Tracks   []TrackRef `json:"tracks"`
}

// Like is one liked album, artist, or playlist. Only the field that matches
// the requested type is set.
type Like struct {
	ID        ID        `json:"id"`
	Timestamp string    `json:"timestamp"`
	Album     *Album    `json:"album"`
	Artist    *Artist   `json:"artist"`
	Playlist  *Playlist `json:"playlist"`
}

// DownloadInfo is one file variant from /tracks/{id}/download-info.
// DownloadInfoURL is the address the API returned. [Client.FileURL] turns
// a plain item into a CDN address. An encraw item stays encrypted.
type DownloadInfo struct {
	Codec           string `json:"codec"`
	BitrateInKbps   int    `json:"bitrateInKbps"`
	Preview         bool   `json:"preview"`
	Direct          bool   `json:"direct"`
	Container       string `json:"container"`
	DownloadInfoURL string `json:"downloadInfoUrl"`
}

// Search is one page of /search. Each *Pager is nil when that type was not
// part of the response. Type "all" does not fill podcasts; pass a concrete
// SearchType for those.
type Search struct {
	Text              string         `json:"text"`
	MisspellCorrected bool           `json:"misspellCorrected"`
	MisspellResult    string         `json:"misspellResult"`
	MisspellOriginal  string         `json:"misspellOriginal"`
	SearchRequestID   string         `json:"searchRequestId"`
	Tracks            *TrackPager    `json:"tracks"`
	Artists           *ArtistPager   `json:"artists"`
	Albums            *AlbumPager    `json:"albums"`
	Playlists         *PlaylistPager `json:"playlists"`
}

// TrackPager is one page of track hits.
type TrackPager struct {
	Total   int     `json:"total"`
	PerPage int     `json:"perPage"`
	Results []Track `json:"results"`
}

// ArtistPager is one page of artist hits.
type ArtistPager struct {
	Total   int      `json:"total"`
	PerPage int      `json:"perPage"`
	Results []Artist `json:"results"`
}

// AlbumPager is one page of album hits.
type AlbumPager struct {
	Total   int     `json:"total"`
	PerPage int     `json:"perPage"`
	Results []Album `json:"results"`
}

// PlaylistPager is one page of playlist hits.
type PlaylistPager struct {
	Total   int        `json:"total"`
	PerPage int        `json:"perPage"`
	Results []Playlist `json:"results"`
}

// Suggestions is the completion list from /search/suggest.
type Suggestions struct {
	Suggestions []string `json:"suggestions"`
}

// ArtistTracks is one page of an artist's tracks.
type ArtistTracks struct {
	Pager  Pager   `json:"pager"`
	Tracks []Track `json:"tracks"`
}

// ArtistAlbums is one page of an artist's albums.
type ArtistAlbums struct {
	Pager  Pager   `json:"pager"`
	Albums []Album `json:"albums"`
}

// Pager is the page block used by artist track and album lists.
type Pager struct {
	Total   int `json:"total"`
	Page    int `json:"page"`
	PerPage int `json:"perPage"`
}

// Genre is one entry from /genres.
type Genre struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	FullTitle     string  `json:"fullTitle"`
	URLPart       string  `json:"urlPart"`
	SubGenres     []Genre `json:"subGenres"`
	ShowInMenu    bool    `json:"showInMenu"`
	ShowInRegions []int   `json:"showInRegions"`
}

// Landing is /landing3. Entities inside a block stay raw because each block
// type (chart, new-releases, personal playlist) uses a different object.
type Landing struct {
	Blocks []LandingBlock `json:"blocks"`
}

// LandingBlock is one landing section.
type LandingBlock struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	TypeForFrom string          `json:"typeForFrom"`
	Title       string          `json:"title"`
	Entities    []LandingEntity `json:"entities"`
}

// LandingEntity is one card inside a landing block. Data is the raw object
// (a track, album, playlist, or chart) so the caller decodes the type it asked for.
type LandingEntity struct {
	ID   ID              `json:"id"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Chart is /landing3/chart. Chart holds the chart playlist.
type Chart struct {
	Chart Playlist `json:"chart"`
}

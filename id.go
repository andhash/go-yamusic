package yamusic

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

// ID is a Yandex Music identifier. The API encodes some ids as JSON numbers
// and others as strings (a track id is often "trackId:albumId"). ID accepts both.
type ID string

// UnmarshalJSON accepts a JSON string or a JSON number.
func (id *ID) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		*id = ""
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*id = ID(s)
		return nil
	}
	*id = ID(string(b))
	return nil
}

// MarshalJSON writes the id as a JSON string.
func (id ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(id))
}

func (id ID) String() string { return string(id) }

// SplitTrackID splits "track:album" into the track id and the album id.
// An id without a colon returns the whole string and an empty album id.
func SplitTrackID(id string) (track, album string) {
	track, album, _ = strings.Cut(id, ":")
	return track, album
}

// joinIDs renders ids the way the batch endpoints expect them:
// one form field, values separated by commas.
func joinIDs(ids []string) (string, error) {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if strings.Contains(id, ",") {
			return "", &Error{Message: "id " + strconv.Quote(id) + " contains a comma"}
		}
		out = append(out, id)
	}
	if len(out) == 0 {
		return "", &Error{Message: "id list is empty"}
	}
	return strings.Join(out, ","), nil
}

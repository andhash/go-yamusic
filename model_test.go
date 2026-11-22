package yamusic

import (
	"encoding/json"
	"testing"
)

func TestTrackLanguage(t *testing.T) {
	var track Track
	if err := json.Unmarshal([]byte(`{"id":"1","title":"A","language":"RU"}`), &track); err != nil || track.Language != "RU" || track.Title != "A" {
		t.Fatalf("%+v %v", track, err)
	}
	track = Track{}
	if err := json.Unmarshal([]byte(`{"id":"1","title":"A"}`), &track); err != nil || track.Language != "" {
		t.Fatalf("missing %+v %v", track, err)
	}
	track = Track{}
	if err := json.Unmarshal([]byte(`{"id":"1","title":"A","language":{"id":"ru"}}`), &track); err != nil || track.Language != "" || track.Title != "A" {
		t.Fatalf("object %+v %v", track, err)
	}
}

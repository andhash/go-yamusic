package yamusic

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestIDAcceptsStringAndNumber(t *testing.T) {
	var id ID
	if err := json.Unmarshal([]byte(`"12:34"`), &id); err != nil {
		t.Fatal(err)
	}
	if id.String() != "12:34" {
		t.Fatalf("string id %q", id)
	}
	if err := json.Unmarshal([]byte(`99`), &id); err != nil {
		t.Fatal(err)
	}
	if id.String() != "99" {
		t.Fatalf("number id %q", id)
	}
	track, album := SplitTrackID("10:20")
	if track != "10" || album != "20" {
		t.Fatalf("split %s %s", track, album)
	}
}

func TestCoverURL(t *testing.T) {
	got := CoverURL("avatars.yandex.net/get-music-content/a/%%", "400x400")
	want := "https://avatars.yandex.net/get-music-content/a/400x400"
	if got != want {
		t.Fatalf("got %s", got)
	}
	if CoverURL("", "200x200") != "" {
		t.Fatal("empty uri")
	}
	abs := "https://cdn.example/img"
	if CoverURL(abs, "200x200") != abs {
		t.Fatal("absolute url was rewritten")
	}
}

func TestTracksRequest(t *testing.T) {
	var gotAuth, gotClient, gotLang, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotClient = r.Header.Get("X-Yandex-Music-Client")
		gotLang = r.Header.Get("Accept-Language")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		if r.URL.Path != "/tracks" || r.Method != http.MethodPost {
			t.Errorf("path %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":[{"id":"7","title":"Yellow","artists":[{"id":1,"name":"Coldplay"}]}]}`))
	}))
	defer srv.Close()

	c := New("secret-token", WithBaseURL(srv.URL), WithLanguage("en"))
	tracks, err := c.Tracks(context.Background(), "7", "8")
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "OAuth secret-token" {
		t.Fatalf("auth %q", gotAuth)
	}
	if gotClient != DefaultClientHeader {
		t.Fatalf("client header %q", gotClient)
	}
	if gotLang != "en" {
		t.Fatalf("lang %q", gotLang)
	}
	vals, err := url.ParseQuery(gotBody)
	if err != nil {
		t.Fatal(err)
	}
	if vals.Get("track-ids") != "7,8" || vals.Get("with-positions") != "true" {
		t.Fatalf("body %q", gotBody)
	}
	if len(tracks) != 1 || tracks[0].Title != "Yellow" || tracks[0].ArtistName() != "Coldplay" {
		t.Fatalf("tracks %+v", tracks)
	}
	if tracks[0].Artists[0].ID.String() != "1" {
		t.Fatal("artist id was not read as a number")
	}
}

func TestGuestOmitsAuthorization(t *testing.T) {
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"result":[]}`))
	}))
	defer srv.Close()
	c := New("", WithBaseURL(srv.URL))
	if _, err := c.Tracks(context.Background(), "1"); err != nil {
		t.Fatal(err)
	}
	if auth != "" {
		t.Fatalf("guest sent %q", auth)
	}
}

func TestHTTPAndEnvelopeErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/account/status":
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":{"name":"unauthorized","message":"token"}}`))
		case "/search":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"error":"validate","error_description":"bad text"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	c := New("t", WithBaseURL(srv.URL))
	_, err := c.AccountStatus(context.Background())
	if !IsUnauthorized(err) {
		t.Fatalf("want unauthorized, got %v", err)
	}
	_, err = c.Search(context.Background(), SearchQuery{Text: "x"})
	var api *Error
	if !errors.As(err, &api) || api.Name != "validate" || api.Message != "bad text" {
		t.Fatalf("envelope error: %v", err)
	}
}

func TestInitFeedsLike(t *testing.T) {
	var likePath, likeBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/account/status":
			_, _ = w.Write([]byte(`{"result":{"account":{"uid":42,"login":"a"},"permissions":{"values":["premium"]}}}`))
		case strings.Contains(r.URL.Path, "/likes/tracks/"):
			likePath = r.URL.Path
			b, _ := io.ReadAll(r.Body)
			likeBody = string(b)
			_, _ = w.Write([]byte(`{"result":{"revision":3}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	c := New("t", WithBaseURL(srv.URL))
	st, err := c.Init(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if c.UID() != "42" || !st.Permissions.Has("premium") {
		t.Fatalf("status %+v uid %s", st, c.UID())
	}
	if err := c.Like(context.Background(), "", "track", "9"); err != nil {
		t.Fatal(err)
	}
	if likePath != "/users/42/likes/tracks/add-multiple" {
		t.Fatalf("path %s", likePath)
	}
	if !strings.Contains(likeBody, "track-ids=9") {
		t.Fatalf("body %s", likeBody)
	}
}

func TestLikeNeedsUser(t *testing.T) {
	c := New("t", WithBaseURL("http://127.0.0.1"))
	err := c.Like(context.Background(), "", "track", "1")
	if err == nil || !strings.Contains(err.Error(), "user id") {
		t.Fatalf("got %v", err)
	}
}

func TestSearchAndSuggest(t *testing.T) {
	var searchQuery, suggestQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/search":
			searchQuery = r.URL.RawQuery
			_, _ = w.Write([]byte(`{"result":{"text":"yellow","tracks":{"total":1,"results":[{"id":"1","title":"Yellow"}]}}}`))
		case "/search/suggest":
			suggestQuery = r.URL.RawQuery
			_, _ = w.Write([]byte(`{"result":{"suggestions":["yellow"]}}`))
		}
	}))
	defer srv.Close()
	c := New("t", WithBaseURL(srv.URL))
	res, err := c.Search(context.Background(), SearchQuery{Text: "yellow", Type: SearchTrack, Page: 1, Nocorrect: true})
	if err != nil {
		t.Fatal(err)
	}
	q, _ := url.ParseQuery(searchQuery)
	if q.Get("text") != "yellow" || q.Get("type") != "track" || q.Get("page") != "1" || q.Get("nocorrect") != "true" {
		t.Fatalf("query %s", searchQuery)
	}
	if res.Tracks == nil || res.Tracks.Results[0].Title != "Yellow" {
		t.Fatalf("search %+v", res)
	}
	sug, err := c.Suggest(context.Background(), "ye")
	if err != nil {
		t.Fatal(err)
	}
	if suggestQuery != "part=ye" || len(sug.Suggestions) != 1 {
		t.Fatalf("suggest %s %+v", suggestQuery, sug)
	}
}

func TestDownloadInfoAndPlaylistDiff(t *testing.T) {
	var diff string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/download-info"):
			_, _ = w.Write([]byte(`{"result":[{"codec":"mp3","bitrateInKbps":320,"preview":false,"downloadInfoUrl":"https://cdn.example/info"}]}`))
		case strings.HasSuffix(r.URL.Path, "/change"):
			_ = r.ParseForm()
			diff = r.Form.Get("diff")
			_, _ = w.Write([]byte(`{"result":{"uid":1,"kind":5,"revision":2,"title":"mine"}}`))
		}
	}))
	defer srv.Close()
	c := New("t", WithBaseURL(srv.URL))
	info, err := c.TrackDownloadInfo(context.Background(), "15")
	if err != nil {
		t.Fatal(err)
	}
	if len(info) != 1 || info[0].BitrateInKbps != 320 || info[0].DownloadInfoURL != "https://cdn.example/info" {
		t.Fatalf("info %+v", info)
	}
	pl, err := c.InsertTracks(context.Background(), "7", 5, 1, 0, []TrackRef{{ID: "15:99"}})
	if err != nil {
		t.Fatal(err)
	}
	if pl.Revision != 2 {
		t.Fatalf("playlist %+v", pl)
	}
	if !strings.Contains(diff, `"op":"insert"`) || !strings.Contains(diff, `"albumId":"99"`) || !strings.Contains(diff, `"at":0`) {
		t.Fatalf("diff %s", diff)
	}
}

func TestChartAndNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/landing3/chart/russia" {
			t.Errorf("path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"name":"not-found","message":"missing"}}`))
	}))
	defer srv.Close()
	c := New("t", WithBaseURL(srv.URL))
	_, err := c.Chart(context.Background(), "russia")
	if !IsNotFound(err) {
		t.Fatalf("got %v", err)
	}
}

func TestIsJSONObjectTrimsSpace(t *testing.T) {
	if !isJSONObject([]byte(" \n{\"a\":1}")) {
		t.Fatal("object")
	}
	if isJSONObject([]byte(" \n[1]")) || isJSONObject(nil) {
		t.Fatal("not an object")
	}
}

func TestEmptyIDs(t *testing.T) {
	c := New("t")
	if _, err := c.Tracks(context.Background()); err == nil {
		t.Fatal("empty id list")
	}
	if _, err := joinIDs([]string{"a,b"}); err == nil {
		t.Fatal("comma in id")
	}
}

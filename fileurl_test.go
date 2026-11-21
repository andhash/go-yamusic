package yamusic

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFileURLSignsJSON(t *testing.T) {
	var sawAuth, sawFormat bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization") == "OAuth secret-token"
		sawFormat = r.URL.Query().Get("format") == "json"
		_, _ = w.Write([]byte(`{"host":"cdn.example","path":"/a/b.mp3","ts":"1700","s":"sig"}`))
	}))
	defer srv.Close()
	c := New("secret-token")
	got, err := c.FileURL(context.Background(), DownloadInfo{DownloadInfoURL: srv.URL + "/info"})
	if err != nil {
		t.Fatal(err)
	}
	if !sawAuth || !sawFormat {
		t.Fatalf("auth %v format %v", sawAuth, sawFormat)
	}
	sum := md5.Sum([]byte(cdnSignPrefix + "a/b.mp3" + "sig"))
	want := "https://cdn.example/get-mp3/" + hex.EncodeToString(sum[:]) + "/1700/a/b.mp3"
	if got != want {
		t.Fatalf("got %s", got)
	}
}

func TestFileURLSignsXMLAndSkipsEncraw(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = w.Write([]byte(`<download-info><host>cdn.example</host><path>/a/b.mp3</path><ts>9</ts><s>z</s></download-info>`))
	}))
	defer srv.Close()
	c := New("t")
	got, err := c.FileURL(context.Background(), DownloadInfo{
		DownloadInfoURL: srv.URL + "/info?format=json",
	})
	if err != nil || !strings.Contains(got, "/get-mp3/") || !strings.HasPrefix(got, "https://cdn.example/") {
		t.Fatalf("%s %v", got, err)
	}
	if _, err := c.FileURL(context.Background(), DownloadInfo{Codec: "encraw", DownloadInfoURL: srv.URL}); err == nil {
		t.Fatal("encraw")
	}
	if hits != 1 {
		t.Fatalf("hits %d", hits)
	}
}

func TestFileURLDirect(t *testing.T) {
	c := New("t")
	got, err := c.FileURL(context.Background(), DownloadInfo{Direct: true, DownloadInfoURL: "https://cdn.example/file.mp3"})
	if err != nil || got != "https://cdn.example/file.mp3" {
		t.Fatalf("%s %v", got, err)
	}
	if _, err := c.FileURL(context.Background(), DownloadInfo{}); err == nil {
		t.Fatal("empty")
	}
}

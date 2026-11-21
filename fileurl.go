package yamusic

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// cdnSignPrefix is the fixed prefix the public web player puts in front of
// the download-info path. It is not an account secret.
const cdnSignPrefix = "XGRlBW9FXlekgbPrRHuSiA"

// FileURL turns one download-info item into an https address of the file.
//
// A direct item returns DownloadInfoURL as-is. Otherwise the client GETs
// DownloadInfoURL (XML or JSON with host, path, ts, and s) and builds the
// same CDN address the public web player uses. The body is not downloaded.
//
// Codec "encraw" is refused. That body is encrypted, and this package does
// not decrypt it. A preview item is not special here: the caller decides
// whether a 30-second clip is acceptable.
func (c *Client) FileURL(ctx context.Context, info DownloadInfo) (string, error) {
	if c == nil {
		return "", &Error{Message: "nil client"}
	}
	if strings.EqualFold(strings.TrimSpace(info.Codec), "encraw") {
		return "", &Error{Name: "encraw", Message: "encrypted stream"}
	}
	if info.Direct {
		if !httpURL(info.DownloadInfoURL) {
			return "", &Error{Message: "direct url is empty"}
		}
		return info.DownloadInfoURL, nil
	}
	rawURL := strings.TrimSpace(info.DownloadInfoURL)
	if rawURL == "" {
		return "", &Error{Message: "download info url is empty"}
	}
	asked, err := withFormatJSON(rawURL)
	if err != nil {
		return "", err
	}
	body, err := c.getRaw(ctx, asked)
	if err != nil {
		return "", err
	}
	loc, err := parseDownloadLocation(body)
	if err != nil {
		return "", err
	}
	return signMP3(loc)
}

type downloadLocation struct {
	Host string
	Path string
	TS   string
	S    string
}

func withFormatJSON(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || !httpURL(raw) {
		return "", &Error{Message: "download info url is invalid"}
	}
	q := u.Query()
	if q.Get("format") == "" {
		q.Set("format", "json")
		u.RawQuery = q.Encode()
	}
	return u.String(), nil
}

func (c *Client) getRaw(ctx context.Context, endpoint string) ([]byte, error) {
	if ctx == nil {
		return nil, &Error{Message: "nil context"}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, &Error{Message: "bad url"}
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json, application/xml, text/xml")
	req.Header.Set("Accept-Language", c.language)
	if c.clientHeader != "" {
		req.Header.Set("X-Yandex-Music-Client", c.clientHeader)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "OAuth "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, &Error{Message: "download info request failed"}
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, &Error{Message: "download info read failed"}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &Error{Status: resp.StatusCode, Message: "download info rejected"}
	}
	return payload, nil
}

func parseDownloadLocation(body []byte) (downloadLocation, error) {
	body = bytesTrim(body)
	if len(body) == 0 {
		return downloadLocation{}, &Error{Message: "download info is empty"}
	}
	if body[0] == '{' {
		var doc struct {
			Host string `json:"host"`
			Path string `json:"path"`
			TS   string `json:"ts"`
			S    string `json:"s"`
		}
		if err := json.Unmarshal(body, &doc); err != nil {
			return downloadLocation{}, &Error{Message: "download info json"}
		}
		return downloadLocation{Host: doc.Host, Path: doc.Path, TS: doc.TS, S: doc.S}, nil
	}
	var doc struct {
		XMLName xml.Name `xml:"download-info"`
		Host    string   `xml:"host"`
		Path    string   `xml:"path"`
		TS      string   `xml:"ts"`
		S       string   `xml:"s"`
	}
	if err := xml.Unmarshal(body, &doc); err != nil || doc.Host == "" {
		return downloadLocation{}, &Error{Message: "download info xml"}
	}
	return downloadLocation{Host: doc.Host, Path: doc.Path, TS: doc.TS, S: doc.S}, nil
}

func signMP3(loc downloadLocation) (string, error) {
	host := strings.TrimSpace(loc.Host)
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	host = strings.Trim(host, "/")
	path := loc.Path
	if host == "" || path == "" || loc.TS == "" || loc.S == "" {
		return "", &Error{Message: "download info is incomplete"}
	}
	// The web player hashes the path without its first character.
	signed := path
	if len(signed) > 0 {
		signed = signed[1:]
	}
	sum := md5.Sum([]byte(cdnSignPrefix + signed + loc.S))
	key := hex.EncodeToString(sum[:])
	return "https://" + host + "/get-mp3/" + key + "/" + loc.TS + path, nil
}

func httpURL(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

func bytesTrim(b []byte) []byte {
	return []byte(strings.TrimSpace(string(b)))
}

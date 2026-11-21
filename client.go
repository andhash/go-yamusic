package yamusic

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	// Version is the library version sent in the User-Agent.
	Version = "0.1.0"

	// DefaultBaseURL is the undocumented Music API host.
	DefaultBaseURL = "https://api.music.yandex.net"

	// DefaultLanguage is the Accept-Language used when the caller does not set one.
	// Known values: en, uz, uk, us, ru, kk, hy.
	DefaultLanguage = "ru"

	// DefaultClientHeader is the client name the API accepts.
	// The server rejects a request that does not look like one of its apps.
	DefaultClientHeader = "YandexMusicAndroid/24023621"
)

// Client talks to the Yandex Music API.
// A Client is safe for concurrent use. The account id stored by Init is guarded.
type Client struct {
	base         string
	token        string
	language     string
	clientHeader string
	userAgent    string
	http         *http.Client

	mu  sync.Mutex
	uid string
}

// Option changes a Client created by New.
type Option func(*Client)

// New builds a client. token is the caller's OAuth token. An empty token
// leaves the Authorization header off, which is the guest session.
func New(token string, opts ...Option) *Client {
	c := &Client{
		base:         DefaultBaseURL,
		token:        strings.TrimSpace(token),
		language:     DefaultLanguage,
		clientHeader: DefaultClientHeader,
		userAgent:    "go-yamusic/" + Version,
		http:         &http.Client{Timeout: 15 * time.Second},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	c.base = strings.TrimRight(c.base, "/")
	return c
}

// WithBaseURL replaces the API host. The value must not have a trailing path
// other than the host root. A trailing slash is stripped.
func WithBaseURL(base string) Option {
	return func(c *Client) {
		if strings.TrimSpace(base) != "" {
			c.base = strings.TrimSpace(base)
		}
	}
}

// WithHTTPClient uses httpClient for every call, including its timeout and transport.
// A nil client is ignored.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.http = httpClient
		}
	}
}

// WithLanguage sets Accept-Language. An empty value keeps the default.
func WithLanguage(lang string) Option {
	return func(c *Client) {
		if strings.TrimSpace(lang) != "" {
			c.language = strings.TrimSpace(lang)
		}
	}
}

// WithClientHeader replaces X-Yandex-Music-Client.
func WithClientHeader(header string) Option {
	return func(c *Client) {
		if strings.TrimSpace(header) != "" {
			c.clientHeader = strings.TrimSpace(header)
		}
	}
}

// WithUserAgent replaces the User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		if strings.TrimSpace(ua) != "" {
			c.userAgent = strings.TrimSpace(ua)
		}
	}
}

// UID returns the account id stored by the last successful Init.
// It is empty until Init runs.
func (c *Client) UID() string {
	if c == nil {
		return ""
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.uid
}

func (c *Client) setUID(id string) {
	c.mu.Lock()
	c.uid = id
	c.mu.Unlock()
}

// user resolves an explicit user id, then the id stored by Init.
func (c *Client) user(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id != "" {
		return id, nil
	}
	if uid := c.UID(); uid != "" {
		return uid, nil
	}
	return "", &Error{Message: "user id is empty; call Init or pass the id"}
}

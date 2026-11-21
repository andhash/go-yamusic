package yamusic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Call sends one API request and returns the unwrapped result.
//
// method is GET or POST. path is absolute ("/tracks") or relative ("tracks").
// query is the URL query. form is the POST body, encoded as
// application/x-www-form-urlencoded. A nil query or form is fine.
//
// The result is the JSON value of the "result" field. When the body has no
// envelope, the whole body is returned. A null result is the JSON literal null.
//
// Call is the escape hatch for an endpoint that has no typed method
// (rotor, queue, pins, presaves, concerts).
func (c *Client) Call(ctx context.Context, method, path string, query, form url.Values) (json.RawMessage, error) {
	if c == nil {
		return nil, &Error{Message: "nil client"}
	}
	if ctx == nil {
		return nil, &Error{Message: "nil context"}
	}
	method = strings.ToUpper(strings.TrimSpace(method))
	if method != http.MethodGet && method != http.MethodPost {
		return nil, &Error{Message: "method must be GET or POST"}
	}

	endpoint, err := c.endpoint(path, query)
	if err != nil {
		return nil, err
	}

	var body io.Reader
	if method == http.MethodPost && len(form) > 0 {
		body = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", c.language)
	if c.clientHeader != "" {
		req.Header.Set("X-Yandex-Music-Client", c.clientHeader)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "OAuth "+c.token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("yamusic: %w", err)
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("yamusic: read response: %w", err)
	}
	return unwrap(resp.StatusCode, payload)
}

func (c *Client) endpoint(path string, query url.Values) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", &Error{Message: "empty path"}
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u, err := url.Parse(c.base + path)
	if err != nil {
		return "", &Error{Message: "bad url: " + err.Error()}
	}
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}
	return u.String(), nil
}

func (c *Client) get(ctx context.Context, path string, query url.Values) (json.RawMessage, error) {
	return c.Call(ctx, http.MethodGet, path, query, nil)
}

func (c *Client) postForm(ctx context.Context, path string, form url.Values) (json.RawMessage, error) {
	return c.Call(ctx, http.MethodPost, path, nil, form)
}

// unwrap reads the Music envelope. A non-2xx status is always *Error.
// A 2xx body whose error field is set is also *Error, with that status.
func unwrap(status int, body []byte) (json.RawMessage, error) {
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		if status >= 200 && status < 300 {
			return json.RawMessage("null"), nil
		}
		return nil, &Error{Status: status, Message: http.StatusText(status)}
	}

	var env struct {
		Result            json.RawMessage `json:"result"`
		Error             json.RawMessage `json:"error"`
		ErrorDescription  string          `json:"error_description"`
		ErrorDescription2 string          `json:"errorDescription"`
		Invocation        json.RawMessage `json:"invocationInfo"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		if status >= 200 && status < 300 {
			return nil, &Error{Status: status, Message: "invalid JSON"}
		}
		return nil, &Error{Status: status, Message: clip(body)}
	}

	name, message := errorFields(env.Error, firstNonEmpty(env.ErrorDescription, env.ErrorDescription2))
	if message != "" || (name != "" && name != "null") {
		if message == "" {
			message = name
		}
		if status < 200 || status >= 300 {
			return nil, &Error{Status: status, Name: name, Message: message}
		}
		return nil, &Error{Status: status, Name: name, Message: message}
	}
	if status < 200 || status >= 300 {
		return nil, &Error{Status: status, Message: http.StatusText(status)}
	}
	if len(env.Result) == 0 || bytes.Equal(bytes.TrimSpace(env.Result), []byte("null")) {
		if len(env.Invocation) == 0 {
			return json.RawMessage(body), nil
		}
		return json.RawMessage("null"), nil
	}
	return env.Result, nil
}

func errorFields(raw json.RawMessage, description string) (name, message string) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return "", strings.TrimSpace(description)
	}
	if raw[0] == '"' {
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return "", clip(raw)
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return "", strings.TrimSpace(description)
		}
		if description != "" {
			return s, description
		}
		return s, s
	}
	var obj struct {
		Name    string `json:"name"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return "", clip(raw)
	}
	name = strings.TrimSpace(obj.Name)
	message = strings.TrimSpace(obj.Message)
	if message == "" {
		message = strings.TrimSpace(description)
	}
	if name == "" && message == "" {
		return "", ""
	}
	return name, message
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func clip(b []byte) string {
	s := string(b)
	if len(s) > 240 {
		return s[:240]
	}
	return s
}

// isJSONObject reports whether raw is a JSON object. Leading space is ignored,
// so a pretty-printed body does not fall through to the list decoder.
func isJSONObject(raw json.RawMessage) bool {
	raw = bytes.TrimSpace(raw)
	return len(raw) > 0 && raw[0] == '{'
}

func decode[T any](raw json.RawMessage) (T, error) {
	var out T
	if len(bytes.TrimSpace(raw)) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return out, nil
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, &Error{Message: "decode " + fmt.Sprintf("%T", out) + ": " + err.Error()}
	}
	return out, nil
}

func decodeList[T any](raw json.RawMessage) ([]T, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return []T{}, nil
	}
	var out []T
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, &Error{Message: "decode list: " + err.Error()}
	}
	if out == nil {
		out = []T{}
	}
	return out, nil
}

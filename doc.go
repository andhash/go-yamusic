// Package yamusic is an unofficial Go client for the Yandex Music HTTP API.
//
// The service does not publish this API. Field names and paths follow the
// same contract as MarshalX/yandex-music-api (https://ym.marshal.dev). A
// response field this package does not model is ignored, so a new server
// field does not break decoding.
//
// The caller supplies an OAuth token. [Client.FileURL] turns a plain
// download-info item into the CDN address of the file. This package does
// not sign /get-file-info or /tracks/{id}/lyrics, and it does not decrypt
// an encraw payload. An endpoint that has no typed method is reachable
// through [Client.Call].
//
// A zero token is a guest session. Guest responses for a full track are
// previews. Methods that need the account id ([Client.Like], playlists)
// use the id stored by [Client.Init] when the caller passes an empty user id.
package yamusic

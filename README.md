# go-yamusic

Неофициальный клиент HTTP API Яндекс Музыки на Go.

Сервис это API не публикует. Набор путей и полей совпадает с контрактом [MarshalX/yandex-music-api](https://github.com/MarshalX/yandex-music-api). Неизвестное поле в ответе игнорируется и разбор не ломает.

Пакет не связан с Яндексом.

## Что есть

Каталог: трек, альбом, альбом с дисками, артист, карточка артиста, треки и альбомы артиста, похожие.

Аккаунт: `Init` запоминает uid, статус, права, настройки, эксперименты.

Плейлисты: список, один плейлист, создание, переименование, видимость, вставка и удаление треков через `diff`.

Лайки и дизлайки треков, альбомов, артистов и плейлистов. Пустой user id берётся из `Init`.

Поиск, подсказки, жанры, лендинг, чарт, новые релизы.

`Client.Call` — любой другой путь (`/rotor/...`, очередь, пины), если типизированного метода нет.

## Чего нет

Токен добывает вызывающий. Пакета для входа нет.

`/get-file-info` и `/tracks/{id}/lyrics` требуют подпись запроса. Этот клиент её не строит.

`TrackDownloadInfo` возвращает то, что отдал API, включая `downloadInfoUrl`. Подписанный CDN-адрес из этого URL клиент не собирает и тело `encraw` не расшифровывает.

Гостевой вызов (пустой токен) получает превью, не полный трек. Полный каталог и Plus зависят от токена аккаунта.

## Установка

```bash
go get github.com/andhash/go-yamusic
```

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/andhash/go-yamusic"
)

func main() {
    c := yamusic.New(os.Getenv("YANDEX_MUSIC_TOKEN"), yamusic.WithLanguage("ru"))
    if _, err := c.Init(context.Background()); err != nil {
        panic(err)
    }
    res, err := c.Search(context.Background(), yamusic.SearchQuery{
        Text: "Boards of Canada",
        Type: yamusic.SearchTrack,
    })
    if err != nil {
        panic(err)
    }
    for _, t := range res.Tracks.Results {
        fmt.Println(t.ArtistName(), "—", t.Title, CoverURL(t))
    }
}

func CoverURL(t yamusic.Track) string {
    return yamusic.CoverURL(t.CoverURI, "400x400")
}
```

Язык ответа: `en`, `uz`, `uk`, `us`, `ru`, `kk`, `hy`. По умолчанию `ru`.

Идентификатор трека в плейлисте — `trackId:albumId`. `SplitTrackID` режет эту пару. `InsertTracks` сам вытаскивает album id, если в `TrackRef` он пустой.

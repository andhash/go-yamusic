# go-yamusic

Клиент HTTP API Яндекс Музыки на Go. Им ищут треки, читают альбомы и плейлисты, ставят лайки и получают адрес файла, который API уже разрешил этому токену.

Сервис это API не публикует. Пути и имена полей те же, что у [MarshalX/yandex-music-api](https://github.com/MarshalX/yandex-music-api). Поле, которого нет в наших структурах, просто пропускается: новое поле на сервере не ломает разбор.

Пакет не связан с Яндексом. Лицензия MIT.

## Установка

```bash
go get github.com/andhash/go-yamusic
```

Нужен Go 1.22 или новее. Зависимостей кроме стандартной библиотеки нет.

## Первый вызов

Токен — это OAuth-строка вашего аккаунта. Библиотека её не добывает: передайте ту, что уже есть. Пустая строка означает гостя. Гость видит каталог, но полный трек ему не отдают, только короткое превью.

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/andhash/go-yamusic"
)

func main() {
    ctx := context.Background()
    c := yamusic.New(os.Getenv("YANDEX_MUSIC_TOKEN"), yamusic.WithLanguage("ru"))

    // Init один раз запоминает uid аккаунта.
    // Лайки и плейлисты с пустым user id берут его отсюда.
    if _, err := c.Init(ctx); err != nil {
        panic(err)
    }

    res, err := c.Search(ctx, yamusic.SearchQuery{
        Text: "Boards of Canada",
        Type: yamusic.SearchTrack,
    })
    if err != nil {
        panic(err)
    }
    for _, t := range res.Tracks.Results {
        fmt.Println(t.ID, t.ArtistName(), "—", t.Title)
        fmt.Println(yamusic.CoverURL(t.CoverURI, "400x400"))
    }
}
```

`New` принимает настройки: другой язык, свой `http.Client` (таймаут, прокси), другой заголовок клиента, другой User-Agent. Таймаут по умолчанию 15 секунд.

Язык ответа: `en`, `uz`, `uk`, `us`, `ru`, `kk`, `hy`. Если не указать, будет `ru`.

## Ошибки

Любая ошибка API имеет тип `*yamusic.Error`. В ней HTTP-статус, короткое имя и текст.

```go
track, err := c.Track(ctx, "123")
if yamusic.IsUnauthorized(err) {
    // токен пустой, протух или не тот
}
if yamusic.IsNotFound(err) {
    // трека нет
}
```

`IsBadRequest` — это 400. Ошибка транспорта (`сеть`, обрыв) завёрнута как обычный `error` и `*yamusic.Error` не является.

Ответ со статусом 2xx, внутри которого сервер положил поле `error`, тоже возвращается как `*yamusic.Error`.

## Идентификаторы

У трека два вида записи.

- `123` — сам трек.
- `123:456` — трек 123 в альбоме 456. В плейлисте и в ссылке с сайта используется эта пара.

`SplitTrackID("123:456")` возвращает `("123", "456")`. `InsertTracks` сам вытаскивает album id из такой пары, если в `TrackRef` поле альбома пустое.

Плейлист адресуется парой `uid:kind`. `uid` — это id пользователя, `kind` — номер плейлиста. «Мне нравится» у пользователя — плейлист с номером 3, но для лайков есть отдельные методы.

`ID` в структурах принимает из JSON и строку, и число, а наружу всегда отдаёт строку. Сравнивать id надо как строки.

## Что умеет клиент

Ниже — обычные задачи. Полный список полей смотрите в godoc пакета `yamusic`.

### Аккаунт

`Init` вызывает `/account/status`, возвращает статус и запоминает uid. `AccountStatus` делает то же без записи uid. `AccountSettings` читает настройки, `SetAccountSettings` пишет одну: имя настройки — это имя поля формы, значение — его содержимое. `Experiments` возвращает сырой JSON экспериментов, как его отдал сервер.

### Поиск

`Search` ищет по тексту. Тип задаётся константой: `SearchAll`, `SearchTrack`, `SearchArtist`, `SearchAlbum`, `SearchPlaylist`, `SearchUser`, `SearchPodcast`, `SearchPodcastEpisode`. Тип `all` подкасты не возвращает — для них нужен явный тип.

`Page` начинается с 0. `Nocorrect: true` отключает исправление опечатки. `PlaylistInBest` просит добавить плейлист в блок «лучшее»; по умолчанию выключено.

`Suggest` — короткие подсказки по началу строки, аргумент `part`.

### Каталог

`Track` и `Tracks` грузят один трек или пачку. Пустой список id — ошибка. Id, внутри которого есть запятая, тоже ошибка: пачка передаётся одной строкой через запятую.

`Album` / `Albums` — альбом без дорожек. `AlbumWithTracks` — альбом с дисками в `Volumes`.

`Artist` / `Artists` — карточка исполнителя. `ArtistBriefInfo` — короткая страница артиста. `ArtistTracks` листает треки (`page`, `pageSize`; ноль означает «не передавать»). `ArtistAlbums` и `ArtistAlsoAlbums` листают альбомы, сортировка `SortByYear` или `SortByRating`. `Similar` и `SimilarArtists` — похожие треки и артисты.

### Плейлисты

Пустой `UserPlaylists` возвращает список плейлистов пользователя. С номерами — только эти номера.

`Playlist` читает один. `CreatePlaylist` создаёт. Пустая видимость становится `private`. Константы: `VisibilityPublic`, `VisibilityPrivate`.

`RenamePlaylist` и `SetPlaylistVisibility` меняют имя и видимость. `DeletePlaylist` удаляет.

Состав меняется одним diff, не по одному треку:

```go
err := c.InsertTracks(ctx, uid, kind, revision, 0,
    yamusic.TrackRef{ID: "123:456"},
)
```

`revision` — текущая ревизия плейлиста. Её берут из самого плейлиста и передают в следующее изменение. `0` в позиции вставки означает начало списка, это не «поле не задано».

`DeleteTrackRange(ctx, uid, kind, revision, from, to)` удаляет полуинтервал `[from, to)`.

### Лайки

`Like`, `Unlike`, `Dislike`, `Undislike`. Тип объекта: `track`, `album`, `artist`, `playlist`. Дизлайк принимает только `track` и `artist`. Пустой user id берётся из `Init`.

Лайк трека снимает дизлайк с этого трека, и наоборот.

Прочитать списки: `LikedTracks`, `LikedAlbums`, `LikedArtists`, `LikedPlaylists`, `DislikedTracks`. У треков в ответе ревизия и пары id, не полные карточки. Чтобы получить карточки, передайте id в `Tracks`.

### Лента

`Landing` — главная. Аргумент `blocks` ограничивает секции, например `"personalplaylists"`, `"new-releases"`, `"chart"`. Пустой список просит набор по умолчанию. Данные блока лежат в `LandingEntity.Data` как сырой JSON: форма у каждой секции своя.

`Chart` — чарт. Пустой аргумент — общий, `"russia"` и другие — региональный. `NewReleases` — новые альбомы. `Genres` — дерево жанров.

### Файл трека

Каталог и файл — разные шаги.

```go
info, err := c.TrackDownloadInfo(ctx, "123:456")
if err != nil {
    panic(err)
}
// Выберите вариант сами. Preview означает клип около 30 секунд.
var picked yamusic.DownloadInfo
for _, item := range info {
    if !item.Preview && item.Codec != "encraw" {
        picked = item
        break
    }
}
addr, err := c.FileURL(ctx, picked)
```

`addr` — обычная https-ссылка на mp3. Скачать байты можно любым HTTP-клиентом. Заголовок OAuth на этот адрес не нужен: право уже зашито в саму ссылку.

`FileURL` не ходит в `/get-file-info` и не расшифровывает `encraw`. Если в списке только такой вариант, полного файла эта библиотека не соберёт. Гость почти всегда получает `Preview: true`.

Обложка к треку собирается отдельно: `CoverURL(track.CoverURI, "400x400")`. Размер — это кусок, которым сервер заменяет `%%` в шаблоне. Подойдут `200x200`, `400x400`, `1000x1000`.

## Своя ручка

Метода на всё нет: радио, очередь, пины, концерты сюда не входили. Их ответ можно забрать сырым JSON:

```go
raw, err := c.Call(ctx, "GET", "/rotor/stations/list", nil, nil)
```

`Call` принимает только GET и POST. Тело POST — форма `application/x-www-form-urlencoded`. Назад приходит уже распакованное поле `result`.

## Чего нет

Входа в аккаунт. Токен приносит вызывающий.

Подписи запросов `/get-file-info` и `/tracks/{id}/lyrics`.

Расшифровки тела `encraw`.

Поэтому библиотека не обещает lossless и не обещает файл там, где сервер отдал только шифрованный поток. Для обычного mp3, который API отдал этому токену, хватает `TrackDownloadInfo` и `FileURL`.

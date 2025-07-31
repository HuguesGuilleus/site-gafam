# Site GAFAM

A scarpper for some GAFAM.

## Usage

```sh
go run main.go list.toml
```

Where `list.toml` is a TOML file like this:

```toml
# You can add multiple group
Group1 = [
  "yt.charts.titles:fr",
  # ...
]
```

- `arte.cat` Arte.tv category
- `arte.ch` Arte.tv channel
- `arte.li` Arte.tv list
- `insta.ch` Instagram channel
- `insta.tr+ch` Instagram and Threads channel
- `lfi.g` actionpopulaire.fr group
- `peertube.a` Peertube account
- `peertube.c` Peertube channel
- `rss` any RSS URL
- `tiktok.ch` TikTok channel 
- `twitch.ch` Twitch channel 
- `twitch.te` Twitch team
- `yt.charts.titles` YouTube charts titles
- `yt.ch` YouTube channel 
- `yt.pl` YouTube playlist

# metatube-sort

A standalone CLI that uses the `metatube-sdk-go` providers to sort a directory of
AV video files / folders by the **first actor's name** (Japanese — see "Limitations"
below). Each actor gets their own folder under the destination directory.

```
src/
  SSIS-001.mp4                       ->  dst/葵つかさ/SSIS-001.mp4
  SNIS-326.mp4                       ->  dst/美里有紗/SNIS-326.mp4
  [Attackers] SHKD-474-C.mp4         ->  dst/<first-actor>/[Attackers] SHKD-474-C.mp4
  [Attackers] SHKD-475/              ->  dst/<first-actor>/[Attackers] SHKD-475/      (folder matched as a unit, NOT recursed into)
```

## Build

```sh
go build -o metatube-sort ./cmd/metatube-sort
```

Cross-compiles cleanly to Windows, macOS, Linux on amd64.

## Usage

```powershell
# PowerShell (Windows) — no configuration needed. The CLI uses the
# same http transport as metatube-server, which honors HTTPS_PROXY
# and the Windows system proxy automatically.
.\metatube-sort.exe -src "D:\movies" -dst "D:\sorted" -db "D:\cache.db" -v
.\metatube-sort.exe -src "D:\movies" -dst "D:\sorted" -db "D:\cache.db" -v --apply
```

```sh
# bash (macOS / Linux)
./metatube-sort -src ./movies -dst ./sorted -db ./cache.db -v
./metatube-sort -src ./movies -dst ./sorted -db ./cache.db -v --apply
```

By default the tool runs in **dry-run** mode: it prints the move plan but
doesn't touch any files. Pass `--apply` to actually move them.

### Flags

| Flag      | Default                | Description                                       |
|-----------|------------------------|---------------------------------------------------|
| `-src`    | (required)             | Source directory to scan                          |
| `-dst`    | (required)             | Destination directory                             |
| `-db`     | `./metatube-sort.db`   | SQLite cache file                                 |
| `-apply`  | `false`                | Actually move files (default is dry-run)          |
| `-j`      | `4`                    | Concurrency                                       |
| `-proxy`  | `""`                   | HTTP/SOCKS5 proxy override (otherwise uses SDK default / env) |
| `-diag`   | `false`                | Network probe only (prints exit info and quits)   |
| `-v`      | `false`                | Verbose logging                                   |

## Behavior

- **Scan**: top-level entries only — no recursion. Files are filtered by
  video extension (`.mp4 .mkv .avi .wmv .ts .mov .m4v .webm .flv .mpg
  .mpeg .rmvb .strm`); directories are kept as-is.
- **Number parsing**: SDK's `common/number.Trim` handles release-group
  brackets, quality suffixes (`-C`, `-UC`, `-HD`, `-4K`), FC2 multi-segment
  IDs, language tags, etc.
- **Resolution order**:
  1. Local SQLite cache (`metatube_sort_alias` table).
  2. JavBus direct HTTP fetch (returns the Japanese actor name).
  3. Gfriends `Filetree.json` — usually a no-op (see Limitations).
  4. Fallback: the original actor name.
- **Multiple actors**: only the first is used.
- **Unmatched**: left untouched (not moved, not deleted).
- **Collisions**: `<name>.mp4` already existing at destination becomes
  `<name>__1.mp4`, `<name>__2.mp4`, ... — the extension is preserved.

## Caching

One table is populated as a side effect:

- `metatube_sort_alias` (this CLI's own) — `number → jp_name → cn_name`.

(The SDK's `movie_metadata` / `actor_metadata` tables are also created
because we wire through `engine.New`, but we don't read or write them.)

To re-resolve from scratch, delete the `-db` file.

## Network requirement (JavBus)

The CLI uses the SDK's own pooled HTTP client
(`hashicorp/go-cleanhttp`'s `DefaultPooledClient`), which sets
`Proxy: http.ProxyFromEnvironment`. This means it honors:

- `HTTPS_PROXY` / `HTTP_PROXY` environment variables
- The Windows system proxy (Settings → Network → Proxy)
- Whatever else Go's stdlib picks up

If `metatube-server` works on the same machine without extra config,
this CLI works too. No special flags required.

If you need to force a specific proxy (e.g. when running under a
non-default port or against a different host):

```powershell
.\metatube-sort.exe -src ... -dst ... -db ... -proxy http://127.0.0.1:7897
```

If you want to confirm the network path is healthy:

```powershell
.\metatube-sort.exe -diag
```

## Limitations

- **Actor names are Japanese, not Chinese.** The SDK does not provide a
  Chinese actor name source out of the box. Folders will be named
  `葵つかさ`, `美里有紗`, etc. — rename them by hand if you want Chinese
  names.
- The Gfriends `Filetree.json` lookup is wired up but yields no Chinese
  names in practice: keys are like `葵つかさ.jpg` and values are studio
  identifiers like `7-S1`, not actor aliases.
- FC2 IDs, Heydouga, and other "special"番号 that JavBus doesn't carry
  will currently fail to resolve. Add more providers via blank imports
  in `internal/enginex/enginex.go` if you need them.
- Network is required for the first run on any new番号; subsequent runs
  hit the local cache.

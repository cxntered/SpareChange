<div align="center">

# [`SpareChange`](https://sparechange.cxntered.dev)

_"changing" Sparebeat maps into osu!mania beatmaps... get it? ahaha..._

</div>

## Showcase

[![Showcase](https://img.youtube.com/vi/a-IGX7RbWXU/maxresdefault.jpg)](https://youtu.be/a-IGX7RbWXU)

## Usage

### Web App

SpareChange is also available as a web app [here](https://sparechange.cxntered.dev)!

### Command Line

```
Usage: sparechange [options] <id>
Options:
  -b, --beta           Whether to fetch a beta Sparebeat map
  -m, --music string   Path to a local .mp3 audio file to use
  -p, --path string    Path to a local Sparebeat map JSON file
```

## Development

### Requirements

- [`Go`](https://go.dev/dl) (`v1.24.5` or higher)
- [`TinyGo`](https://tinygo.org/getting-started/install): For building the WebAssembly module, _optional_ (`v0.39.0` or higher)
  - [`binaryen`](https://github.com/WebAssembly/binaryen): Required on Windows for WebAssembly builds

### Building

#### Command Line

```bash
$ go build -o bin/SpareChange ./cmd/cli
```

#### Web App

```bash
# Copy the images folder to the web directory
$ cp -r internal/assets/images web/

# Build the WebAssembly module
$ GOOS=js GOARCH=wasm tinygo build -o web/main.wasm -no-debug ./cmd/wasm
```

## Resources

- [Sparebeat](https://sparebeat.com) ([beta version](https://beta.sparebeat.com))
- [osu!](https://osu.ppy.sh)
- [Sparebeat Map Editor](https://github.com/bo-yakitarako/sparebeat-map-editor)
- [Sparebeat Map Format](/docs/sparebeat-maps.md)
- [.osu file format specification](<https://osu.ppy.sh/wiki/en/Client/File_formats/osu_(file_format)>)
- [osu-parser](https://github.com/Waffle-osu/osu-parser)

## To Do

_not comprehensive at all, just a rough list for now_

- [x] Fetch Sparebeat map from both beta and stable
- [x] Parse Sparebeat map JSON
- [x] Convert Sparebeat map to an osu!mania beatmap
- [x] Create a `.osz` file with the converted map and audio
- [x] Properly convert Sparebeat BPM & speed changes to osu!mania SV
- [x] Allow local Sparebeat maps to be converted
- [ ] Allow osu!mania beatmaps to be converted into Sparebeat maps

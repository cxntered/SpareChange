## Sparebeat Map Format

The Sparebeat map format is a JSON object with the following structure:

```json
{
    "id": string,        // Unique map identifier (not present in beta)
    "title": string,     // Title of the song
    "artist": string,    // Artist of the song
    "url": string,       // Author-provided URL
    "bgColor": string[], // Background colors in hex (optional)
    "beats": number,     // Number of beats in the song (optional, defaults to 4(?))
    "bpm": number,       // Beats per minute
    "startTime": number, // Start time of the song (in milliseconds)

    // Level difficulty rating OR level name
    "level": {
        "easy": number | string,
        "normal": number | string,
        "hard": number | string
    },

    // Map data for each difficulty
    "map": {
        "easy": Array<string | object>,
        "normal": Array<string | object>,
        "hard": Array<string | object>
    }
}
```

Each level's difficulty rating or name is set by the map author, with seemingly no restrictions. In practice, no levels are named and instead use an arbitrary(?) difficulty rating. In beta, setting a level to 0 will gray it out, while setting it to -1 or a string will hide it entirely.

Each difficulty's map data is an array of strings and/or objects. Strings contain the note data of each section, where a section is 4 measures long. Objects contain map settings, which include toggling on or off the bar line, changing the BPM, and changing the speed of the map.

### Section Format

A section string may look like this, where each row is separated by a comma:

```
23,,(a3),4,,e,,,[67]
```

Numbers `1`-`4` represent normal (4<sup>th</sup>) notes, where 1 is in the leftmost column and 4 is in the rightmost column.

`a`-`d` represent hold note starts, while `e`-`h` represent hold note ends.

Numbers `5`-`8` represent attack notes, i.e. the red notes where hitting them while pressing the spacebar at the same time will reward double points.

Parentheses `()` represent 24<sup>th</sup> notes, where any measures inside the parentheses will use notes that are 1/6 of a beat long.

Square brackets `[]` represent bind zones, where any rows inside the brackets will prevent "ghost tapping" (pressing keys when there are no notes). If there is no closing bracket, the bind zone will persist.

### Map Options Format

Map options are objects with the following structure:

```json
{
    "barLine": boolean, // Whether to show the bar line
    "bpm": number,      // Sets the BPM from that point onwards
    "speed": number     // Sets the speed from that point onwards
}
```

## Fetching a map's data

### Beta

`beta.sparebeat.com/play/[map ID]` is the standard URL for Sparebeat maps in beta. You can find the map ID in the URL when viewing a map. The map IDs seem to random 8 character long hexadecimal strings.

Using that ID, you can fetch the map's data by sending a GET request to `beta.sparebeat.com/api/tracks/[map ID]/map`. This will return the JSON data for the map.

To fetch the song audio, `beta.sparebeat.com/api/tracks/[map ID]/audio` will return the audio file as an MP3 file.

### Stable

`sparebeat.com/play/[map ID]` is the standard URL for Sparebeat maps in stable. The map IDs in stable seem to be author-generated and are not random like in beta.

You can fetch the map's data by sending a GET request to `sparebeat.com/play/[map ID]/map`. This will return the JSON data for the map. Stable map data seems identical to beta, with the exception of the map ID being present in stable

To fetch the song audio, `sparebeat.com/play/[map ID]/music` will return the audio file as an MP3 file.

## Notes

This was heavily assisted by the [Sparebeat Map Editor](https://github.com/bo-yakitarako/sparebeat-map-editor), which provides [TypeScript typings](https://github.com/bo-yakitarako/sparebeat-map-editor/blob/master/src/modules/mapConvert/ISparebeatJson.ts) for the Sparebeat map format. To my knowledge, the map format is not documented anywhere else, so through that repo and a bit of testing by myself, I think this is a good enough description of the Sparebeat map format, though it may not be 100% accurate and wording might not be the best.

As well, [this video](https://www.youtube.com/watch?v=Jgz8-UKv8NE) was a blessing as it shows the different types of notes in rhythm games, which I was losing my mind over because god I hate music theory.

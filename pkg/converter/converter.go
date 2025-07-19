package converter

import (
	"encoding/json"
	"strconv"
	"strings"
	"unicode"

	"github.com/cxntered/SpareChange/pkg/types"
)

func ConvertSparebeatToOsu(sbMap types.SparebeatMap) (types.OsuMap, error) {
	var osuMap types.OsuMap

	osuMap.General = types.GeneralSection{
		AudioFilename: "audio.mp3",
		Mode:          types.ModeMania,
	}

	osuMap.Metadata = types.MetadataSection{
		Title:         sbMap.Title,
		TitleUnicode:  sbMap.Title,
		Artist:        sbMap.Artist,
		ArtistUnicode: sbMap.Artist,
		Creator:       "Sparebeat",
		Source:        sbMap.URL,
	}

	// TODO: placeholder values, probably change later
	osuMap.Difficulty = types.DifficultySection{
		HPDrainRate:       5,
		CircleSize:        4,
		OverallDifficulty: 5,
		ApproachRate:      5,
		SliderMultiplier:  1.4,
		SliderTickRate:    1,
	}

	osuMap.Events.List = []types.Event{
		{
			EventType: types.EventTypeBackground,
			StartTime: sbMap.StartTime,
			EventParams: types.EventParams{
				FileName: "background.png",
				XOffset:  0,
				YOffset:  0,
			},
		},
	}

	var beats uint = 4
	if sbMap.Beats != 0 {
		beats = sbMap.Beats
	}
	bpm := getBPM(sbMap.BPM)
	osuMap.BPM = types.TimingPoint{
		Time:        sbMap.StartTime,
		BeatLength:  60 * 1000 / bpm,
		Meter:       beats,
		SampleSet:   1,
		SampleIndex: 0,
		Volume:      100,
		Uninherited: true,
		Effects:     types.EffectNone,
	}

	if isLevelEnabled(sbMap.Level.Easy) {
		easy, err := convertSparebeatDifficulty(sbMap, osuMap, "Easy")
		if err != nil {
			return osuMap, err
		}
		osuMap.Difficulties = append(osuMap.Difficulties, easy)
	}

	if isLevelEnabled(sbMap.Level.Normal) {
		normal, err := convertSparebeatDifficulty(sbMap, osuMap, "Normal")
		if err != nil {
			return osuMap, err
		}
		osuMap.Difficulties = append(osuMap.Difficulties, normal)
	}

	if isLevelEnabled(sbMap.Level.Hard) {
		hard, err := convertSparebeatDifficulty(sbMap, osuMap, "Hard")
		if err != nil {
			return osuMap, err
		}
		osuMap.Difficulties = append(osuMap.Difficulties, hard)
	}

	return osuMap, nil
}

func convertSparebeatDifficulty(sbMap types.SparebeatMap, osuMap types.OsuMap, levelName string) (types.OsuFile, error) {
	var osuFile types.OsuFile

	osuFile.Version = 14
	osuFile.General = osuMap.General
	osuFile.Metadata = osuMap.Metadata
	osuFile.Metadata.Version = levelName
	osuFile.Difficulty = osuMap.Difficulty
	osuFile.Events = osuMap.Events
	osuFile.TimingPoints.List = append(osuFile.TimingPoints.List, osuMap.BPM)

	var mapData []interface{}
	switch levelName {
	case "Easy":
		mapData = sbMap.Map.Easy
	case "Normal":
		mapData = sbMap.Map.Normal
	case "Hard":
		mapData = sbMap.Map.Hard
	default:
		mapData = sbMap.Map.Hard
	}

	var rowCount uint = 0
	bpm := getBPM(sbMap.BPM)
	var beats uint = 4
	if sbMap.Beats != 0 {
		beats = sbMap.Beats
	}
	prevBeats := beats
	holdNotes := make(map[uint]int) // column index -> start time
	in24thMode := false

	for _, elem := range mapData {
		switch v := elem.(type) {
		case string:
			hitObjects := parseSections(v, sbMap, &rowCount, &bpm, &beats, prevBeats, holdNotes, in24thMode)
			osuFile.HitObjects.List = append(osuFile.HitObjects.List, hitObjects...)

		case map[string]interface{}:
			timingPoint := parseMapOptions(v, sbMap, rowCount, &bpm, beats)
			if timingPoint != (types.TimingPoint{}) {
				osuFile.TimingPoints.List = append(osuFile.TimingPoints.List, timingPoint)
			}
		}
	}

	return osuFile, nil
}

func parseSections(
	section string,
	sbMap types.SparebeatMap,
	rowCount *uint,
	bpm *float64,
	beats *uint,
	prevBeats uint,
	holdNotes map[uint]int,
	in24thMode bool,
) []types.HitObject {
	rows := strings.Split(section, ",")
	beatLength := 60 * 1000 / *bpm

	var hitObjects []types.HitObject
	for _, row := range rows {
		*rowCount++
		time := int(float64(*rowCount**beats)*beatLength/16) - max(sbMap.StartTime, -sbMap.StartTime)
		notes := strings.SplitSeq(row, "")

		for note := range notes {
			if isNumeric(note) { // normal notes
				lane, _ := strconv.Atoi(note)
				if lane > 4 { // convert attack notes into normal notes
					lane -= 4
				}

				hitObjects = append(hitObjects, types.HitObject{
					XPosition: int16((512 * lane / 4) - 64),
					YPosition: 192,
					Time:      time,
					Type:      types.HitCircle,
					HitSound:  types.HitSoundNormal,
					HitSample: types.HitSample{
						NormalSet:   0,
						AdditionSet: 0,
						Index:       0,
						Volume:      0,
					},
				})
			} else if unicode.IsLetter(rune(note[0])) { // hold notes
				// convert letter into alphabet index (i.e. lane)
				lane := uint(unicode.ToLower(rune(note[0]))) - uint('a') + 1

				if lane <= 4 {
					holdNotes[lane] = time
				} else if lane <= 8 {
					lane -= 4
					startTime, ok := holdNotes[lane]

					if ok {
						hitObjects = append(hitObjects, types.HitObject{
							XPosition: int16((512 * lane / 4) - 64),
							YPosition: 192,
							Time:      startTime,
							Type:      types.HoldNote,
							HitSound:  100,
							ObjectParams: types.ObjectParams{
								EndTime: time,
							},
							HitSample: types.HitSample{
								NormalSet:   0,
								AdditionSet: 0,
								Index:       0,
								Volume:      0,
							},
						})
						delete(holdNotes, lane)
					}
				}
			} else { // modifiers
				if note == "(" {
					in24thMode = true
					*beats = 6
					break
				} else if note == ")" {
					in24thMode = false
					*beats = prevBeats
					break
				}
			}
		}
	}

	return hitObjects
}

func parseMapOptions(
	mapOptions map[string]interface{},
	sbMap types.SparebeatMap,
	rowCount uint,
	bpm *float64,
	beats uint,
) types.TimingPoint {
	var opts types.MapOptions
	mapBytes, _ := json.Marshal(mapOptions)
	if err := json.Unmarshal(mapBytes, &opts); err == nil {
		if opts.BPM != nil {
			*bpm = *opts.BPM
		}
		beatLength := 60 * 1000 / *bpm

		time := sbMap.StartTime + int(float64(rowCount*beats)*beatLength/16)

		if opts.BPM != nil {
			return types.TimingPoint{
				Time:        time,
				BeatLength:  beatLength,
				Meter:       beats,
				SampleSet:   0,
				SampleIndex: 0,
				Volume:      100,
				Uninherited: true,
				Effects:     0,
			}
		} else if opts.Speed != nil {
			return types.TimingPoint{
				Time:        time,
				BeatLength:  -100 / *opts.Speed,
				Meter:       beats,
				SampleSet:   0,
				SampleIndex: 0,
				Volume:      100,
				Uninherited: false,
				Effects:     0,
			}
		}
	}

	return types.TimingPoint{}
}

func isLevelEnabled(val interface{}) bool {
	if val == nil {
		return false
	}
	switch v := val.(type) {
	case float64:
		return v > 0
	case string:
		return isNumeric(v) && v != "0" && v != "-1"
	default:
		return false
	}
}

func getBPM(bpmRaw interface{}) float64 {
	if bpmRaw == nil {
		return 0
	}
	switch v := bpmRaw.(type) {
	case float64:
		return v
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return f
		}
		return 0
	default:
		return 0
	}
}

func isNumeric(str string) bool {
	for _, char := range str {
		if !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}

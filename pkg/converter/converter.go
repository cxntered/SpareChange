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

	easy, err := convertSparebeatDifficulty(sbMap, osuMap, "Easy")
	if err != nil {
		return osuMap, err
	}
	normal, err := convertSparebeatDifficulty(sbMap, osuMap, "Normal")
	if err != nil {
		return osuMap, err
	}
	hard, err := convertSparebeatDifficulty(sbMap, osuMap, "Hard")
	if err != nil {
		return osuMap, err
	}

	osuMap.Difficulties = []types.OsuFile{easy, normal, hard}

	return osuMap, nil
}

func convertSparebeatDifficulty(sbMap types.SparebeatMap, osuMap types.OsuMap, levelName string) (types.OsuFile, error) {
	var osuFile types.OsuFile

	osuFile.Version = 14
	osuFile.General = osuMap.General
	osuFile.Metadata = osuMap.Metadata
	osuFile.Metadata.Version = levelName
	osuFile.Difficulty = osuMap.Difficulty

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
	var bpm float64 = sbMap.BPM
	var beats uint
	if sbMap.Beats != 0 {
		beats = sbMap.Beats
	} else {
		beats = 4
	}
	holdNotes := make(map[uint]int) // column index -> start time

	for _, elem := range mapData {
		switch v := elem.(type) {
		case string:
			hitObjects := parseSections(v, sbMap, &rowCount, &bpm, beats, holdNotes)
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

func parseSections(section string, sbMap types.SparebeatMap, rowCount *uint, bpm *float64, beats uint, holdNotes map[uint]int) []types.HitObject {
	rows := strings.Split(section, ",")
	beatLength := 60 * 1000 / *bpm

	var hitObjects []types.HitObject
	for _, row := range rows {
		*rowCount++
		time := int(float64(*rowCount*beats)*beatLength/16) - sbMap.StartTime
		notes := strings.SplitSeq(row, "")

		// TODO: handle 24th notes & bind zones
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
			} else { // hold notes
				// convert letter into alphabet index (i.e. lane)
				lane := uint(unicode.ToLower(rune(note[0]))) - uint('a') + 1

				if lane <= 4 {
					holdNotes[lane] = time
				} else {
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
			}
		}
	}

	return hitObjects
}

func parseMapOptions(mapOptions map[string]interface{}, sbMap types.SparebeatMap, rowCount uint, bpm *float64, beats uint) types.TimingPoint {
	var opts types.MapOptions
	mapBytes, _ := json.Marshal(mapOptions)
	if err := json.Unmarshal(mapBytes, &opts); err == nil {
		if opts.BPM != nil {
			*bpm = *opts.BPM
		}
		beatLength := 60 * 1000 / *bpm

		time := int(float64(rowCount*beats)*beatLength/16) - sbMap.StartTime

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

func isNumeric(str string) bool {
	for _, char := range str {
		if !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}

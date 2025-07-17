package converter

import (
	"encoding/json"

	"github.com/cxntered/SpareChange/pkg/types"
)

func ConvertSparebeatToOsu(sbMap types.SparebeatMap) (types.OsuMap, error) {
	var osuMap types.OsuMap

	osuMap.General = types.GeneralSection{
		AudioFilename: "audio.mp3",
		Mode:          3,
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

	var bpm float64 = sbMap.BPM
	var rowCount uint = 0

	for _, elem := range mapData {
		switch v := elem.(type) {
		case string:
			// TODO: parse rows
			rowCount++

		case map[string]interface{}:
			timingPoint := parseMapOptions(v, sbMap, &bpm, rowCount)
			if timingPoint != (types.TimingPoint{}) {
				osuFile.TimingPoints.TimingPoints = append(osuFile.TimingPoints.TimingPoints, timingPoint)
			}
		}
	}

	return osuFile, nil
}

func parseMapOptions(mapOptions map[string]interface{}, sbMap types.SparebeatMap, bpm *float64, rowCount uint) types.TimingPoint {
	var opts types.MapOptions
	mapBytes, _ := json.Marshal(mapOptions)
	if err := json.Unmarshal(mapBytes, &opts); err == nil {
		if opts.BPM != nil {
			*bpm = *opts.BPM
		}
		beatLength := 60 * 1000 / *bpm

		var beats uint
		if sbMap.Beats != 0 {
			beats = sbMap.Beats
		} else {
			beats = 4
		}

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

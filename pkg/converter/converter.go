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

	for i, elem := range mapData {
		switch v := elem.(type) {
		case string:
			// TODO: parse rows

		case map[string]interface{}:
			var opts types.MapOptions
			mapBytes, _ := json.Marshal(v)
			if err := json.Unmarshal(mapBytes, &opts); err == nil {
				// TODO: current bpm can change, need to track it somewhere
				bpm := sbMap.BPM
				if opts.BPM != nil {
					bpm = *opts.BPM
				}
				beatLength := 60 * 1000 / bpm

				var beats int
				if sbMap.Beats != 0 {
					beats = sbMap.Beats
				} else {
					beats = 4
				}
				time := int((i + 1) * beats * int(beatLength))

				if opts.BPM != nil {
					osuFile.TimingPoints.TimingPoints = append(osuFile.TimingPoints.TimingPoints, types.TimingPoint{
						Time:        time,
						BeatLength:  beatLength,
						Meter:       beats,
						SampleSet:   0,
						SampleIndex: 0,
						Volume:      100,
						Uninherited: true,
						Effects:     0,
					})
				} else if opts.Speed != nil {
					osuFile.TimingPoints.TimingPoints = append(osuFile.TimingPoints.TimingPoints, types.TimingPoint{
						Time:        time,
						BeatLength:  -100 / *opts.Speed,
						Meter:       beats,
						SampleSet:   0,
						SampleIndex: 0,
						Volume:      100,
						Uninherited: false,
						Effects:     0,
					})
				}
			}
		}
	}

	return osuFile, nil
}

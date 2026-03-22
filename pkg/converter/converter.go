package converter

import (
	"encoding/json"
	"strconv"
	"strings"
	"unicode"

	"github.com/cxntered/SpareChange/pkg/types"
)

func ConvertSparebeatToOsu(sbMap types.SparebeatMap) (types.OsuBeatmap, error) {
	var osuMap types.OsuBeatmap

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

	if isLevelEnabled(sbMap.Level.Easy.Value) {
		easy, err := convertSparebeatDifficulty(sbMap, osuMap, "Easy")
		if err != nil {
			return osuMap, err
		}
		osuMap.Difficulties = append(osuMap.Difficulties, easy)
	}

	if isLevelEnabled(sbMap.Level.Normal.Value) {
		normal, err := convertSparebeatDifficulty(sbMap, osuMap, "Normal")
		if err != nil {
			return osuMap, err
		}
		osuMap.Difficulties = append(osuMap.Difficulties, normal)
	}

	if isLevelEnabled(sbMap.Level.Hard.Value) {
		hard, err := convertSparebeatDifficulty(sbMap, osuMap, "Hard")
		if err != nil {
			return osuMap, err
		}
		osuMap.Difficulties = append(osuMap.Difficulties, hard)
	}

	return osuMap, nil
}

func convertSparebeatDifficulty(sbMap types.SparebeatMap, osuMap types.OsuBeatmap, levelName string) (types.OsuDifficulty, error) {
	var osuDiff types.OsuDifficulty

	osuDiff.Version = 14
	osuDiff.General = osuMap.General
	osuDiff.Metadata = osuMap.Metadata
	osuDiff.Metadata.Version = levelName
	osuDiff.Difficulty = osuMap.Difficulty
	osuDiff.Events = osuMap.Events

	var mapData []types.MapEntry
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

	var elapsedTime float64 = 0
	bpm := getBPM(sbMap.BPM.Value)
	baseBPM := bpm
	var beats uint = 4
	if sbMap.Beats != 0 {
		beats = sbMap.Beats
	}
	prevBeats := beats
	holdNotes := make(map[uint]int) // column index -> start time
	in24thMode := false
	inBindZone := false

	for _, elem := range mapData {
		switch {
		case elem.Section != "":
			hitObjects, timingPoints := parseSections(elem.Section, sbMap.StartTime, &elapsedTime, bpm, &beats, prevBeats, holdNotes, &in24thMode, &inBindZone)
			osuDiff.HitObjects.List = append(osuDiff.HitObjects.List, hitObjects...)
			osuDiff.TimingPoints.List = append(osuDiff.TimingPoints.List, timingPoints...)

		case elem.Options != nil:
			timingPoints, err := parseMapOptions(elem.Options, sbMap.StartTime, elapsedTime, &bpm, beats, baseBPM)
			if err != nil {
				return osuDiff, err
			}
			for _, timingPoint := range timingPoints {
				if timingPoint.Time < sbMap.StartTime && len(osuDiff.TimingPoints.List) == 0 {
					osuDiff.TimingPoints.List = append(osuDiff.TimingPoints.List, types.TimingPoint{
						Time:        0,
						BeatLength:  60 * 1000 / bpm,
						Meter:       beats,
						SampleSet:   0,
						SampleIndex: 0,
						Volume:      100,
						Uninherited: true,
						Effects:     types.EffectOmitFirstBarLine,
					})
				}
				osuDiff.TimingPoints.List = append(osuDiff.TimingPoints.List, timingPoint)
			}
		}
	}

	osuDiff.TimingPoints.List = append(osuDiff.TimingPoints.List, types.TimingPoint{
		Time:        sbMap.StartTime,
		BeatLength:  60 * 1000 / bpm,
		Meter:       beats,
		SampleSet:   0,
		SampleIndex: 0,
		Volume:      100,
		Uninherited: true,
		Effects:     types.EffectNone,
	})

	return osuDiff, nil
}

func parseSections(
	section string,
	startTime int,
	elapsedTime *float64,
	bpm float64,
	beats *uint,
	prevBeats uint,
	holdNotes map[uint]int,
	in24thMode *bool,
	inBindZone *bool,
) ([]types.HitObject, []types.TimingPoint) {
	rows := strings.Split(section, ",")
	beatLength := 60 * 1000 / bpm
	var hitObjects []types.HitObject
	var timingPoints []types.TimingPoint

	for _, row := range rows {
		time := startTime + int(*elapsedTime) - int(beatLength/4)
		notes := strings.SplitSeq(row, "")

		for note := range notes {
			if unicode.IsDigit(rune(note[0])) { // normal notes
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
							HitSound:  types.HitSoundNormal,
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
				if note == "(" && !*in24thMode {
					*in24thMode = true
					*beats = 6
					continue
				} else if note == ")" && *in24thMode {
					*in24thMode = false
					*beats = prevBeats
					continue
				} else if note == "[" && !*inBindZone {
					*inBindZone = true
					timingPoints = append(timingPoints, types.TimingPoint{
						Time:        time,
						BeatLength:  -100,
						Meter:       *beats,
						SampleSet:   0,
						SampleIndex: 0,
						Volume:      100,
						Uninherited: false,
						Effects:     types.EffectKiaiTime,
					})
					continue
				} else if note == "]" && *inBindZone {
					*inBindZone = false
					timingPoints = append(timingPoints, types.TimingPoint{
						Time:        time,
						BeatLength:  -100,
						Meter:       *beats,
						SampleSet:   0,
						SampleIndex: 0,
						Volume:      100,
						Uninherited: false,
						Effects:     types.EffectNone,
					})
					continue
				}
			}
		}

		*elapsedTime += beatLength / float64(*beats)
	}

	return hitObjects, timingPoints
}

func parseMapOptions(
	mapOptions json.RawMessage,
	startTime int,
	elapsedTime float64,
	bpm *float64,
	beats uint,
	baseBPM float64,
) ([]types.TimingPoint, error) {
	opts, err := types.ParseMapOptions(mapOptions)
	if err != nil {
		return nil, err
	}
	if opts.BPM != nil {
		if *opts.BPM == 0 {
			*bpm = 0.000001 // bpm cannot be zero, so we set it to a small value instead
		} else {
			*bpm = *opts.BPM
		}
	}
	beatLength := 60 * 1000 / *bpm
	time := startTime + int(elapsedTime) - int(beatLength/4)

	if opts.BPM != nil {
		return []types.TimingPoint{
			{
				Time:        time,
				BeatLength:  beatLength,
				Meter:       beats,
				SampleSet:   0,
				SampleIndex: 0,
				Volume:      100,
				Uninherited: true,
				Effects:     types.EffectNone,
			},
			{
				Time:        time,
				BeatLength:  -100 / (baseBPM / *bpm), // keep scroll speed relative to base BPM
				Meter:       beats,
				SampleSet:   0,
				SampleIndex: 0,
				Volume:      100,
				Uninherited: false,
				Effects:     types.EffectNone,
			},
		}, nil
	} else if opts.Speed != nil {
		var speed float64
		if *opts.Speed == 0 {
			speed = 0.000001 // speed cannot be zero, so we set it to a small value instead
		} else {
			speed = *opts.Speed
		}
		return []types.TimingPoint{{
			Time:        time,
			BeatLength:  -100 / speed,
			Meter:       beats,
			SampleSet:   0,
			SampleIndex: 0,
			Volume:      100,
			Uninherited: false,
			Effects:     types.EffectNone,
		}}, nil
	}

	return []types.TimingPoint{}, nil
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

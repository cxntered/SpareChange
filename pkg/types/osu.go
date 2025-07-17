package types

// many sections can be shared between different difficulties
type OsuMap struct {
	General      GeneralSection
	Metadata     MetadataSection
	Difficulty   DifficultySection
	Difficulties []OsuFile
}

type OsuFile struct {
	Version int8

	General      GeneralSection
	Editor       EditorSection
	Metadata     MetadataSection
	Difficulty   DifficultySection
	Events       EventsSection
	TimingPoints TimingPointsSection
	Colours      ColoursSection
	HitObjects   HitObjectsSection
}

type GeneralSection struct {
	AudioFilename            string
	AudioLeadIn              int
	AudioHash                string
	PreviewTime              int
	Countdown                uint8   // 0: no countdown, 1: normal, 2: half, 3: double, defaults to 1
	SampleSet                string  // "Normal", "Soft", or "Drum", defaults to "Normal"
	StackLeniency            float32 // decimal between 0 and 1, defaults to 0.7
	Mode                     uint8   // 0: osu!standard, 1: osu!taiko, 2: osu!catch, 3: osu!mania
	LetterboxInBreaks        bool
	StoryFireInFront         bool
	UseSkinSprites           bool
	AlwaysShowPlayfield      bool
	OverlayPosition          string // "NoChange", "Below", or "Above", defaults to "NoChange"
	SkinPreference           string
	EpilepsyWarning          bool
	CountdownOffset          int
	SpecialStyle             bool
	WidescreenStoryboard     bool
	SamplesMatchPlaybackRate bool
}

type EditorSection struct {
	Bookmarks       []int
	DistanceSpacing float64
	BeatDivisor     int
	GridSize        int
	TimelineZoom    float64
}

type MetadataSection struct {
	Title         string
	TitleUnicode  string
	Artist        string
	ArtistUnicode string
	Creator       string
	Version       string
	Source        string
	Tags          []string
	BeatmapID     int
	BeatmapSetID  int
}

type DifficultySection struct {
	HPDrainRate       float64
	CircleSize        float64
	OverallDifficulty float64
	ApproachRate      float64
	SliderMultiplier  float64
	SliderTickRate    float64
}

type EventsSection struct {
	Events []Event
}

type Event struct {
	EventType   interface{} // can be string or int
	StartTime   int
	EventParams EventParams
}

type EventParams struct {
	// background & video specific
	FileName string
	XOffset  int16
	YOffset  int16

	// break specific
	EndTime int
}

type TimingPointsSection struct {
	TimingPoints []TimingPoint
}

type TimingPoint struct {
	Time        int
	BeatLength  float64
	Meter       uint
	SampleSet   int
	SampleIndex int
	Volume      int
	Uninherited bool
	Effects     int
}

type ColoursSection struct {
	Colours []Colour
}

type Colour struct {
	Option string
	Red    uint8
	Green  uint8
	Blue   uint8
	Alpha  uint8
}

type HitObjectsSection struct {
	HitObjects []HitObject
}

type HitObject struct {
	XPosition    int16
	YPosition    int16
	Time         int
	HitSound     int8
	ObjectParams ObjectParams
	HitSample    []HitSample
}

type ObjectParams struct {
	// slider specific
	CurveType   rune
	CurvePoints []string
	Slides      int
	Length      float64
	EdgeSounds  []int8
	EdgeSets    []string

	// spinner & hold specific
	EndTime int
}

type HitSample struct {
	NormalSet   uint8
	AdditionSet uint8
	Index       int
	Volume      int
	FileName    string
}

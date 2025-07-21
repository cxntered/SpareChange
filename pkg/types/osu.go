package types

// many sections can be shared between different difficulties
type OsuMap struct {
	General      GeneralSection
	Metadata     MetadataSection
	Difficulty   DifficultySection
	Events       EventsSection
	TimingPoints TimingPointsSection
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
	Countdown                Countdown
	SampleSet                SampleSet
	StackLeniency            float32 // decimal between 0 and 1, defaults to 0.7
	Mode                     Mode
	LetterboxInBreaks        bool
	StoryFireInFront         bool
	UseSkinSprites           bool
	AlwaysShowPlayfield      bool
	OverlayPosition          OverlayPosition
	SkinPreference           string
	EpilepsyWarning          bool
	CountdownOffset          int
	SpecialStyle             bool
	WidescreenStoryboard     bool
	SamplesMatchPlaybackRate bool
}

type Countdown uint8

const (
	CountdownNoChange Countdown = 0
	CountdownNormal   Countdown = 1
	CountdownHalf     Countdown = 2
	CountdownDouble   Countdown = 3
)

type Mode uint8

const (
	ModeStandard Mode = 0
	ModeTaiko    Mode = 1
	ModeCatch    Mode = 2
	ModeMania    Mode = 3
)

type SampleSet string

const (
	SampleSetNormal SampleSet = "Normal"
	SampleSetSoft   SampleSet = "Soft"
	SampleSetDrum   SampleSet = "Drum"
)

type OverlayPosition string

const (
	OverlayPositionNoChange OverlayPosition = "NoChange"
	OverlayPositionBelow    OverlayPosition = "Below"
	OverlayPositionAbove    OverlayPosition = "Above"
)

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
	List []Event
}

type Event struct {
	EventType   EventType
	StartTime   int
	EventParams EventParams
}

type EventType uint8

const (
	EventTypeBackground EventType = 0
	EventTypeVideo      EventType = 1
	EventTypeBreak      EventType = 2
)

type EventParams struct {
	// background & video specific
	FileName string
	XOffset  int16
	YOffset  int16

	// break specific
	EndTime int
}

type TimingPointsSection struct {
	List []TimingPoint
}

type TimingPoint struct {
	Time        int
	BeatLength  float64
	Meter       uint
	SampleSet   int
	SampleIndex int
	Volume      int
	Uninherited bool
	Effects     Effect
}

type Effect uint8

const (
	EffectNone             Effect = 1
	EffectKiaiTime         Effect = 1 << 0
	EffectOmitFirstBarLine Effect = 1 << 3
)

type ColoursSection struct {
	List []Colour
}

type Colour struct {
	Option string
	Red    uint8
	Green  uint8
	Blue   uint8
	Alpha  uint8
}

type HitObjectsSection struct {
	List []HitObject
}

type HitObject struct {
	XPosition    int16
	YPosition    int16
	Time         int
	Type         HitObjectType
	HitSound     HitSound
	ObjectParams ObjectParams
	HitSample    HitSample
}

type HitSound uint8

const (
	HitSoundNormal  HitSound = 1 << 0
	HitSoundWhistle HitSound = 1 << 1
	HitSoundFinish  HitSound = 1 << 2
	HitSoundClap    HitSound = 1 << 3
)

type HitObjectType uint8

const (
	HitCircle HitObjectType = 1 << 0
	Slider    HitObjectType = 1 << 1
	NewCombo  HitObjectType = 1 << 2
	Spinner   HitObjectType = 1 << 3
	HoldNote  HitObjectType = 1 << 7
)

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

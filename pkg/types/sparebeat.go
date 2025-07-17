package types

type SparebeatMap struct {
	ID        string   `json:"id,omitempty"`
	Title     string   `json:"title"`
	Artist    string   `json:"artist"`
	URL       string   `json:"url"`
	BgColor   []string `json:"bgColor,omitempty"`
	Beats     int      `json:"beats,omitempty"`
	BPM       float64  `json:"bpm"`
	StartTime int      `json:"startTime"`
	Level     Level    `json:"level"`
	Map       MapData  `json:"map"`
}

type Level struct {
	Easy   interface{} `json:"easy"`
	Normal interface{} `json:"normal"`
	Hard   interface{} `json:"hard"`
}

type MapData struct {
	Easy   []interface{} `json:"easy"`
	Normal []interface{} `json:"normal"`
	Hard   []interface{} `json:"hard"`
}

type MapOptions struct {
	BarLine *bool    `json:"barLine,omitempty"`
	BPM     *float64 `json:"bpm,omitempty"`
	Speed   *float64 `json:"speed,omitempty"`
}

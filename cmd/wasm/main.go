package main

import (
	"bytes"
	"encoding/json"
	"syscall/js"

	"github.com/cxntered/SpareChange/pkg/converter"
	"github.com/cxntered/SpareChange/pkg/types"
	"github.com/cxntered/SpareChange/pkg/utils"
)

func main() {
	js.Global().Set("convertSparebeatMap", js.FuncOf(convertSparebeatMap))
	<-make(chan struct{}) // keep program running
}

func convertSparebeatMap(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return map[string]interface{}{
			"success": false,
			"error":   "No map data provided",
		}
	}

	var sbMap types.SparebeatMap
	err := json.Unmarshal([]byte(args[0].String()), &sbMap)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Invalid map data: " + err.Error(),
		}
	}

	osuMap, err := converter.ConvertSparebeatToOsu(sbMap)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Conversion error: " + err.Error(),
		}
	}

	files := make(map[string]interface{})
	for _, diff := range osuMap.Difficulties {
		var buf bytes.Buffer
		err := converter.WriteOsuContent(diff, &buf)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error":   "Failed to generate .osu file: " + err.Error(),
			}
		}

		fileName := diff.Metadata.Artist + " - " + diff.Metadata.Title + " (" + diff.Metadata.Creator + ") [" + diff.Metadata.Version + "].osu"
		files[utils.Sanitize(fileName)] = buf.String()
	}

	return map[string]interface{}{
		"success": true,
		"metadata": map[string]interface{}{
			"title":  osuMap.Metadata.Title,
			"artist": osuMap.Metadata.Artist,
		},
		"files": files,
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cxntered/SpareChange/pkg/converter"
	"github.com/cxntered/SpareChange/pkg/types"
	flag "github.com/spf13/pflag"
)

func main() {
	beta := flag.BoolP("beta", "b", false, "Convert a beta Sparebeat map")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Usage: sparechange [options] <id>")
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	id := args[0]
	var mapURL string
	if *beta {
		mapURL = fmt.Sprintf("https://beta.sparebeat.com/api/tracks/%s/map", id)
	} else {
		mapURL = fmt.Sprintf("https://sparebeat.com/play/%s/map", id)
	}

	fmt.Printf("Fetching Sparebeat map from: %s\n", mapURL)

	res, err := http.Get(mapURL)
	if err != nil {
		fmt.Printf("Error fetching map: %v\n", err)
		os.Exit(1)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		os.Exit(1)
	}

	var mapData types.SparebeatMap
	err = json.Unmarshal(body, &mapData)
	if err != nil {
		fmt.Printf("Error unmarshalling response body: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully fetched map: %+v\n", mapData.Title)

	osuMap, err := converter.ConvertSparebeatToOsu(mapData)
	if err != nil {
		fmt.Printf("Error converting Sparebeat map to osu! format: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully converted map to osu! format: %+v\n", osuMap.Difficulties[len(osuMap.Difficulties)-1])
}

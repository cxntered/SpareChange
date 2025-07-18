package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

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

	var mapURL string = fmt.Sprintf("https://sparebeat.com/play/%s/map", id)
	if *beta {
		mapURL = fmt.Sprintf("https://beta.sparebeat.com/api/tracks/%s/map", id)
	}
	fmt.Printf("Fetching Sparebeat map from: %s\n", mapURL)

	// fetch the map data
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

	// parse the map data
	var mapData types.SparebeatMap
	err = json.Unmarshal(body, &mapData)
	if err != nil {
		fmt.Printf("Error unmarshalling response body: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Fetched & parsed map: %+v\n", mapData.Title)

	// convert map to osu! format
	osuMap, err := converter.ConvertSparebeatToOsu(mapData)
	if err != nil {
		fmt.Printf("Error converting Sparebeat map to osu! format: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Converted map to osu! format")

	// create temp dir for conversion
	tempDir := os.TempDir()
	sparebeatDir := filepath.Join(tempDir, "sparechange")
	err = os.MkdirAll(sparebeatDir, os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating temp directory: %v\n", err)
		os.Exit(1)
	}

	// write osu! files for each difficulty
	var diffFiles []string
	for _, diffMap := range osuMap.Difficulties {
		fileName := fmt.Sprintf("%s - %s (%s) [%s].osu",
			diffMap.Metadata.Artist,
			diffMap.Metadata.Title,
			diffMap.Metadata.Creator,
			diffMap.Metadata.Version,
		)
		file := filepath.Join(sparebeatDir, fileName)
		diffFiles = append(diffFiles, file)
		err = converter.WriteOsuFile(diffMap, file)
		if err != nil {
			fmt.Printf("Error writing osu! file: %v\n", err)
			os.Exit(1)
		}
	}
	fmt.Println("Wrote map difficulties' .osu files")

	// download audio file
	var audioURL string = fmt.Sprintf("https://sparebeat.com/play/%s/music", id)
	if *beta {
		audioURL = fmt.Sprintf("https://beta.sparebeat.com/api/tracks/%s/audio", id)
	}
	audioFile := filepath.Join(sparebeatDir, "audio.mp3")

	resp, err := http.Get(audioURL)
	if err != nil {
		fmt.Printf("Error downloading audio file: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	out, err := os.Create(audioFile)
	if err != nil {
		fmt.Printf("Error creating audio file: %v\n", err)
		os.Exit(1)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("Error saving audio file: %v\n", err)
		os.Exit(1)
	}
	out.Close()
	fmt.Println("Downloaded music audio file")

	// zip all files into a .osz
	files := append(diffFiles, audioFile)
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current working directory: %v\n", err)
		os.Exit(1)
	}

	err = converter.ZipFiles(files, filepath.Join(cwd, fmt.Sprintf("%s - %s.osz", osuMap.Metadata.Artist, osuMap.Metadata.Title)))
	if err != nil {
		fmt.Printf("Error zipping files: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created .osz file: %s - %s.osz\n", osuMap.Metadata.Artist, osuMap.Metadata.Title)

	// clean up temp directory
	err = os.RemoveAll(sparebeatDir)
	if err != nil {
		fmt.Printf("Error cleaning up temp folder: %v\n", err)
	} else {
		fmt.Println("Cleaned up temporary conversion folder")
	}
}

package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cxntered/SpareChange/pkg/converter"
	"github.com/cxntered/SpareChange/pkg/types"
	"github.com/cxntered/SpareChange/pkg/utils"
	"github.com/disintegration/imaging"
	flag "github.com/spf13/pflag"
)

//go:embed assets/background.png
var background []byte

func main() {
	beta := flag.BoolP("beta", "b", false, "Whether to fetch a beta Sparebeat map")
	path := flag.StringP("path", "p", "", "Path to a local Sparebeat map JSON file")
	music := flag.StringP("music", "m", "", "Path to a local .mp3 audio file to use")
	flag.Parse()

	args := flag.Args()
	if *path == "" && len(args) == 0 {
		fmt.Println("Usage: sparechange [options] <id>")
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var sbMap types.SparebeatMap

	if *path != "" {
		file, err := os.Open(*path)
		if err != nil {
			fmt.Printf("Error opening map file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		body, err := io.ReadAll(file)
		if err != nil {
			fmt.Printf("Error reading map file: %v\n", err)
			os.Exit(1)
		}
		err = json.Unmarshal(body, &sbMap)
		if err != nil {
			fmt.Printf("Error parsing map JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Parsed local map: %+v\n", sbMap.Title)
	} else {
		id := args[0]
		var mapURL string = fmt.Sprintf("https://sparebeat.com/play/%s/map", id)
		if *beta {
			mapURL = fmt.Sprintf("https://beta.sparebeat.com/api/tracks/%s/map", id)
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
		err = json.Unmarshal(body, &sbMap)
		if err != nil {
			fmt.Printf("Error unmarshalling response body: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Fetched & parsed map: %+v\n", sbMap.Title)
	}

	// convert map to osu! format
	osuMap, err := converter.ConvertSparebeatToOsu(sbMap)
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

	// handle music
	var audioFile string = filepath.Join(sparebeatDir, "audio.mp3")
	if *music != "" {
		in, err := os.Open(*music)
		if err != nil {
			fmt.Printf("Error opening music file: %v\n", err)
			os.Exit(1)
		}
		defer in.Close()
		out, err := os.Create(audioFile)
		if err != nil {
			fmt.Printf("Error creating audio file: %v\n", err)
			os.Exit(1)
		}
		_, err = io.Copy(out, in)
		if err != nil {
			fmt.Printf("Error copying audio file: %v\n", err)
			os.Exit(1)
		}
		out.Close()
		fmt.Println("Copied local music audio file")
	} else {
		id := ""
		if *path == "" && len(args) > 0 {
			id = args[0]
		}
		var audioURL string = fmt.Sprintf("https://sparebeat.com/play/%s/music", id)
		if *beta {
			audioURL = fmt.Sprintf("https://beta.sparebeat.com/api/tracks/%s/audio", id)
		}
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
	}

	// create background image
	img, _, err := image.Decode(bytes.NewReader(background))
	if err != nil {
		fmt.Printf("Error decoding background image: %v\n", err)
		os.Exit(1)
	}
	width, height := img.Bounds().Dx(), img.Bounds().Dy()

	gradient := imaging.New(width, height, color.Transparent)
	startColor := color.NRGBA{R: 67, G: 198, B: 172, A: 255}
	endColor := color.NRGBA{R: 25, G: 22, B: 84, A: 255}
	if len(sbMap.BgColor) == 2 {
		startColor = utils.HexToNRGBA(sbMap.BgColor[0])
		endColor = utils.HexToNRGBA(sbMap.BgColor[1])
	}
	for y := range height {
		t := float64(y) / float64(height-1)
		c := utils.InterpolateColor(startColor, endColor, t)
		draw.Draw(gradient, image.Rect(0, y, width, y+1), image.NewUniform(c), image.Point{}, draw.Over)
	}

	blended := imaging.Overlay(img, gradient, image.Pt(0, 0), 0.8)

	backgroundPath := filepath.Join(sparebeatDir, "background.png")
	imaging.Save(blended, backgroundPath)
	fmt.Println("Created background image")

	// zip all files into a .osz
	files := append(diffFiles, audioFile, backgroundPath)

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

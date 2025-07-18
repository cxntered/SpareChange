package converter

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cxntered/SpareChange/pkg/types"
)

func WriteOsuFile(osuFile types.OsuFile, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("osu file format v%d\n", osuFile.Version))

	// general
	sb.WriteString("[General]\n")
	sb.WriteString(fmt.Sprintf("AudioFilename: %s\n", osuFile.General.AudioFilename))
	sb.WriteString(fmt.Sprintf("Mode: %d\n", osuFile.General.Mode))
	sb.WriteString("\n")

	// metadata
	sb.WriteString("[Metadata]\n")
	sb.WriteString(fmt.Sprintf("Title: %s\n", osuFile.Metadata.Title))
	sb.WriteString(fmt.Sprintf("TitleUnicode: %s\n", osuFile.Metadata.TitleUnicode))
	sb.WriteString(fmt.Sprintf("Artist: %s\n", osuFile.Metadata.Artist))
	sb.WriteString(fmt.Sprintf("ArtistUnicode: %s\n", osuFile.Metadata.ArtistUnicode))
	sb.WriteString(fmt.Sprintf("Creator: %s\n", osuFile.Metadata.Creator))
	sb.WriteString(fmt.Sprintf("Version: %s\n", osuFile.Metadata.Version))
	sb.WriteString(fmt.Sprintf("Source: %s\n", osuFile.Metadata.Source))
	sb.WriteString("\n")

	// difficulty
	sb.WriteString("[Difficulty]\n")
	sb.WriteString(fmt.Sprintf("HPDrainRate: %.1f\n", osuFile.Difficulty.HPDrainRate))
	sb.WriteString(fmt.Sprintf("CircleSize: %.1f\n", osuFile.Difficulty.CircleSize))
	sb.WriteString(fmt.Sprintf("OverallDifficulty: %.1f\n", osuFile.Difficulty.OverallDifficulty))
	sb.WriteString(fmt.Sprintf("ApproachRate: %.1f\n", osuFile.Difficulty.ApproachRate))
	sb.WriteString(fmt.Sprintf("SliderMultiplier: %.1f\n", osuFile.Difficulty.SliderMultiplier))
	sb.WriteString(fmt.Sprintf("SliderTickRate: %.1f\n", osuFile.Difficulty.SliderTickRate))
	sb.WriteString("\n")

	// timing points
	sb.WriteString("[TimingPoints]\n")
	for _, timingPoint := range osuFile.TimingPoints.List {
		var uninherited uint8 = 0
		if timingPoint.Uninherited {
			uninherited = 1
		}

		sb.WriteString(fmt.Sprintf("%d,%.2f,%d,%d,%d,%d,%d\n",
			timingPoint.Time,
			timingPoint.BeatLength,
			timingPoint.Meter,
			timingPoint.SampleSet,
			timingPoint.SampleIndex,
			timingPoint.Volume,
			uninherited,
		))
	}
	sb.WriteString("\n")

	// hitobjects
	sb.WriteString("[HitObjects]\n")
	for _, hitObject := range osuFile.HitObjects.List {
		hitSample := fmt.Sprintf("%d:%d:%d:%d:",
			hitObject.HitSample.NormalSet,
			hitObject.HitSample.AdditionSet,
			hitObject.HitSample.Index,
			hitObject.HitSample.Volume,
		)

		switch hitObject.Type {
		case types.HitCircle:
			sb.WriteString(fmt.Sprintf("%d,%d,%d,%d,%d,%s\n",
				hitObject.XPosition,
				hitObject.YPosition,
				hitObject.Time,
				hitObject.Type,
				hitObject.HitSound,
				hitSample,
			))
		case types.HoldNote:
			sb.WriteString(fmt.Sprintf("%d,%d,%d,%d,%d,%d:%s\n",
				hitObject.XPosition,
				hitObject.YPosition,
				hitObject.Time,
				hitObject.Type,
				hitObject.HitSound,
				hitObject.ObjectParams.EndTime,
				hitSample,
			))
		}
	}

	_, err = f.WriteString(sb.String())
	return err
}

func ZipFiles(files []string, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.Base(file)
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, f)
		if err != nil {
			return err
		}
	}

	return nil
}

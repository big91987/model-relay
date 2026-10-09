package main

import (
	"encoding/json"
	"fmt"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type imageFact struct {
	Path    string `json:"path"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Uniform bool   `json:"uniform"`
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: audit-png <repository-relative-evidence-directory>")
		os.Exit(2)
	}
	facts := []imageFact{}
	err := filepath.WalkDir(os.Args[1], func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() || !strings.HasSuffix(path, ".png") {
			return walkErr
		}
		file, openErr := os.Open(path)
		if openErr != nil {
			return openErr
		}
		image, decodeErr := png.Decode(file)
		file.Close()
		if decodeErr != nil {
			return decodeErr
		}
		bounds := image.Bounds()
		firstRed, firstGreen, firstBlue, firstAlpha := image.At(bounds.Min.X, bounds.Min.Y).RGBA()
		uniform := true
		for vertical := bounds.Min.Y; vertical < bounds.Max.Y && uniform; vertical++ {
			for horizontal := bounds.Min.X; horizontal < bounds.Max.X; horizontal++ {
				red, green, blue, alpha := image.At(horizontal, vertical).RGBA()
				if red != firstRed || green != firstGreen || blue != firstBlue || alpha != firstAlpha {
					uniform = false
					break
				}
			}
		}
		facts = append(facts, imageFact{Path: filepath.ToSlash(path), Width: bounds.Dx(), Height: bounds.Dy(), Uniform: uniform})
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "PNG audit failed")
		os.Exit(1)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(facts); err != nil {
		os.Exit(1)
	}
}

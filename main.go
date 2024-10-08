package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Tnze/go-mc/nbt"
)

func main() {
	defaultOutputLocation := "output"

	var inputFolder string
	var outputFolder string
	flag.StringVar(&inputFolder, "i", "", "the full link to the input folder")
	flag.StringVar(&outputFolder, "o", defaultOutputLocation, "the name of the output folder")
	flag.Parse()

	flag.VisitAll(func(f *flag.Flag) {
		if f.Value.String() == "" {
			fmt.Println(f.Name, "not set (see -h)")
			os.Exit(0)
		}
	})
	
	// read files in input folder
	entries, err := os.ReadDir(inputFolder)
	if err != nil {
		log.Fatal(err)
	}
	
	outputFolderFullPath, err := resolvePath(outputFolder)
	if err != nil {
		log.Fatal(err)
	}
	
	if outputFolder == defaultOutputLocation {
		// choose path for subfolder using world name
		worldName, err := getWorldName(inputFolder)
		outputFolderFullPath = filepath.Join(outputFolderFullPath, worldName)
		if err != nil {
			log.Fatal(err)
		}
	
		// make the subfolder
		err = createFolderIfNotExist(outputFolderFullPath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// no subfolders or creating folders
		// just ensure write access
		if _, err := os.Stat(outputFolder); err != nil {
			log.Fatal("Could not find the folder: ", outputFolder)
		}
	}

	startTime := time.Now()
	allColors := createAllColors()

	var wg sync.WaitGroup
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "map_") && strings.HasSuffix(e.Name(), ".dat") {
			wg.Add(1)
			go func(entry os.DirEntry) {
				defer wg.Done()
				// read data
				filePath := filepath.Join(inputFolder, entry.Name())

				mapdata := openFile(filePath)
				data := mapdata["data"].(map[string]interface{})
				colors := data["colors"].([]uint8)

				// create image
				var pixels []Pixel
				for _, c := range colors {
					pixels = append(pixels, allColors[c])
				}

				img := createImageFromPixels(pixels)

				// save the image to a file
				outputFileName := entry.Name()[:len(entry.Name())-4] + ".png"
				outputFileLocation := filepath.Join(outputFolderFullPath, outputFileName)
				outputFile, err := os.Create(outputFileLocation)
				if err != nil {
					fmt.Println("Error creating output file:", err)
					return
				}
				defer outputFile.Close()

				err = png.Encode(outputFile, img)
				if err != nil {
					fmt.Println("Error encoding PNG:", err)
					return
				}
			}(e)
		}
	}
	wg.Wait()
	elapsedTime := time.Since(startTime)
	fmt.Println("exported", len(entries), "maps in", elapsedTime)
	
	fmt.Println("output saved to", outputFolderFullPath)
}

func resolvePath(inputPath string) (string, error) {
	// convert the input path to an absolute path
	absPath, err := filepath.Abs(inputPath)
	if err != nil {
		return "", fmt.Errorf("error getting absolute path: %w", err)
	}

	return absPath, nil
}

func getWorldName(absPath string) (string, error) {
	// Clean and split the path into components
	cleanPath := filepath.Clean(absPath)
	sections := strings.Split(cleanPath, string(filepath.Separator))

	// Ensure there are at least two sections in the path
	if len(sections) < 2 {
		return "", fmt.Errorf("path does not have enough sections: %s", absPath)
	}

	// Return the second-last section
	return sections[len(sections)-2], nil
}

func createFolderIfNotExist(folderPath string) error {
	// Check if the folder exists
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		// Create the folder if it does not exist
		err := os.MkdirAll(folderPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create folder: %w", err)
		}
		fmt.Printf("folder created: %s\n", folderPath)
	} else {
		fmt.Printf("folder already exists: %s\n", folderPath)
	}
	return nil
}

type Pixel [4]uint8

func createAllColors() []Pixel {
	var baseColors = []Pixel{
		{0, 0, 0, 0},
		{127, 178, 56, 255},
		{247, 233, 163, 255},
		{199, 199, 199, 255},
		{255, 0, 0, 255},
		{160, 160, 255, 255},
		{167, 167, 167, 255},
		{0, 124, 0, 255},
		{255, 255, 255, 255},
		{164, 168, 184, 255},
		{151, 109, 77, 255},
		{112, 112, 112, 255},
		{64, 64, 255, 255},
		{143, 119, 72, 255},
		{255, 252, 245, 255},
		{216, 127, 51, 255},
		{178, 76, 216, 255},
		{102, 153, 216, 255},
		{229, 229, 51, 255},
		{127, 204, 25, 255},
		{242, 127, 165, 255},
		{76, 76, 76, 255},
		{153, 153, 153, 255},
		{76, 127, 153, 255},
		{127, 63, 178, 255},
		{51, 76, 178, 255},
		{102, 76, 51, 255},
		{102, 127, 51, 255},
		{153, 51, 51, 255},
		{25, 25, 25, 255},
		{250, 238, 77, 255},
		{92, 219, 213, 255},
		{74, 128, 255, 255},
		{0, 217, 58, 255},
		{129, 86, 49, 255},
		{112, 2, 0, 255},
		{209, 177, 161, 255},
		{159, 82, 36, 255},
		{149, 87, 108, 255},
		{112, 108, 138, 255},
		{186, 133, 36, 255},
		{103, 117, 53, 255},
		{160, 77, 78, 255},
		{57, 41, 35, 255},
		{135, 107, 98, 255},
		{87, 92, 92, 255},
		{122, 73, 88, 255},
		{76, 62, 92, 255},
		{76, 50, 35, 255},
		{76, 82, 42, 255},
		{142, 60, 46, 255},
		{37, 22, 16, 255},
		{189, 48, 49, 255},
		{148, 63, 97, 255},
		{92, 25, 29, 255},
		{22, 126, 134, 255},
		{58, 142, 140, 255},
		{86, 44, 62, 255},
		{20, 180, 133, 255},
		{100, 100, 100, 255},
		{216, 175, 147, 255},
		{127, 167, 150, 255},
	}
	var multipliers = []uint8{
		180, 220, 255, 135,
	}
	var allColors []Pixel
	for _, color := range baseColors {
		for _, multiplier := range multipliers {
			allColors = append(allColors, multiplyColor(color, multiplier))
		}
	}
	return allColors
}

func multiplyColor(inputPixel Pixel, multiplier uint8) Pixel {
	var newPixel Pixel
	for i, channelValue := range inputPixel {
		if i < 3 {
			newPixel[i] = uint8(math.Floor(float64(channelValue) * float64(multiplier) / 255.0))
		} else {
			newPixel[i] = inputPixel[i]
		}
	}
	return newPixel
}

func createImageFromPixels(pixels []Pixel) *image.RGBA {
	sideLength := int(math.Sqrt(float64(len(pixels))))
	img := image.NewRGBA(image.Rect(0, 0, sideLength, sideLength))

	for y := 0; y < sideLength; y++ {
		for x := 0; x < sideLength; x++ {
			index := y*sideLength + x
			img.Set(x, y, color.RGBA{R: pixels[index][0], G: pixels[index][1], B: pixels[index][2], A: pixels[index][3]})
		}
	}

	return img
}

type MapData map[string]interface{}

func openFile(filepath string) MapData {
	b, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer
	err = gunzipWrite(&buf, b)
	if err != nil {
		log.Fatal(err)
	}

	var mapdata MapData
	err = nbt.Unmarshal(buf.Bytes(), &mapdata)
	if err != nil {
		log.Fatal(err)
	}

	return mapdata
}

func gunzipWrite(w io.Writer, data []byte) error {
	gr, err := gzip.NewReader(bytes.NewBuffer(data))
	defer gr.Close()
	data, err = io.ReadAll(gr)
	if err != nil {
		return err
	}
	w.Write(data)
	return nil
}

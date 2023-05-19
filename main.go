package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	exif "github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
)

type Metadata struct {
	Path                string
	Latitude, Longitude string
}

var files []string
var (
	exifTagIndex   = exif.NewTagIndex()
	exifIfdMapping = exifcommon.NewIfdMapping()
)
var meta []Metadata

func main() {
	csvFile := flag.String("csv", "output.csv", "output CSV file")

	// Parse command-line flags e.g -csv filename.csv
	flag.Parse()

	// Create CSV file for writing
	f, err := os.Create(*csvFile)
	if err != nil {
		log.Fatalf("failed to create CSV file: %v", err)
	}
	defer f.Close()

	// Create CSV writer
	writer := csv.NewWriter(f)
	defer writer.Flush()

	// Writing colmn headers in csv file
	if err := writer.Write([]string{"path", "latitude", "longitude"}); err != nil {
		log.Fatalf("failed to write CSV header: %v", err)
	}

	// Getting all the images from /images and its subdirectories
	err = filepath.Walk("images", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return nil
		}

		if !info.IsDir() && isSupportedImageExt(path) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		log.Fatalf("failed to walk through dir & sub-dir : %v", err)
	}

	// Loading IFD's for mapping
	if err := exifcommon.LoadStandardIfds(exifIfdMapping); err != nil {
		panic(err)
	}

	for _, file := range files {
		if isSupportedImageExt(file) {

			// Read each image file
			fileBytes, err := ioutil.ReadFile(file)
			if err != nil {
				log.Printf("failed to read image file %s: %v", file, err)
				continue
			}

			// Searches for an EXIF blob in the byte-slice.
			rawExif, err := exif.SearchAndExtractExif(fileBytes)
			if err != nil {
				log.Printf("%s: %v", file, err)
				continue
			}

			// Getting all exifTags & IFD's from exif returns exif index
			_, ifdIndex, err := exif.Collect(exifIfdMapping, exifTagIndex, rawExif)
			if err != nil {
				log.Printf("%s: %v", file, err)
				continue
			}

			ifd, err := ifdIndex.RootIfd.ChildWithIfdPath(exifcommon.IfdGpsInfoStandardIfdIdentity)
			if err != nil {
				log.Printf("%s: %v", file, err)
				continue
			}

			// Getting Gps Information (Lat Long)
			gi, err := ifd.GpsInfo()
			if err != nil {
				log.Printf("failed to get GPS info: %s", err)
				continue
			}

			latitudeStr := strconv.FormatFloat(gi.Latitude.Decimal(), 'f', -1, 64)
			longitudeStr := strconv.FormatFloat(gi.Longitude.Decimal(), 'f', -1, 64)

			meta = append(meta, Metadata{
				Path:      file,
				Latitude:  latitudeStr,
				Longitude: longitudeStr,
			})

			// Write row to CSV
			if err := writer.Write([]string{file, latitudeStr, longitudeStr}); err != nil {
				log.Printf("failed to write CSV row for file %s: %v", file, err)
				continue
			}
		}
	}
	fmt.Printf("CSV file '%s' generated successfully.\n", *csvFile)

	// Extra Credit Work :-)
	GenerateHTML(*csvFile)
}

func GenerateHTML(csvFile string) {
	htmlTemplate := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Image EXIF Data</title>
		</head>
		<body>
			<table>
				<tr>
					<th>Path</th>
					<th>Latitude</th>
					<th>Longitude</th>
				</tr>
				{{range .}}
				<tr>
					<td><img src='{{.Path}}' /></td>
					<td>{{.Latitude}}</td>
					<td>{{.Longitude}}</td>
				</tr>
				{{end}}
			</table>
		</body>
		</html>
	`

	// Create a new template and parse the HTML template string
	tmpl := template.Must(template.New("metadataTable").Parse(htmlTemplate))

	// Create a new file to write the HTML output (filename same as csv file)
	file, err := os.Create(strings.TrimSuffix(csvFile, ".csv") + ".html")
	if err != nil {
		fmt.Printf("Failed to create file: %v", err)
		return
	}
	defer file.Close()

	// Execute the template with the data and write the result to the file
	err = tmpl.Execute(file, meta)
	if err != nil {
		fmt.Printf("Failed to write HTML: %v", err)
		return
	}

	fmt.Printf("HTML file '%s' generated successfully.\n", file.Name())
}

func isSupportedImageExt(file string) bool {
	ext := filepath.Ext(file)
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif"
}

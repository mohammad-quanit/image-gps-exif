package main

import (
	"encoding/csv"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func Test_isSupportedImageExt(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Check jpg format ", args{file: "image.jpg"}, true},
		{"Check jpeg format ", args{file: "image.jpeg"}, true},
		{"Check png format ", args{file: "image.png"}, true},
		{"Check gif format ", args{file: "image.gif"}, true},
		{"Check pdf format ", args{file: "image.pdf"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSupportedImageExt(tt.args.file); got != tt.want {
				t.Errorf("isSupportedImageExt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Main(t *testing.T) {
	// Set up test environment
	tmpCSVFile := "output.csv"
	tmpHTMLFile := "output.html"
	imagesDir := "images"

	// Clean up test files after the test completes
	defer func() {
		os.Remove(tmpCSVFile)
		os.Remove(tmpHTMLFile)
	}()

	// Create test files
	createTestCSV(tmpCSVFile)
	createTestImages(imagesDir)

	// Set command-line arguments for the test
	os.Args = []string{"main", "-csv", tmpCSVFile}

	// Run the main function
	main()

	// Verify the CSV file is generated correctly
	verifyCSVFile(t, tmpCSVFile)
}

// Helper function to create a test CSV file
func createTestCSV(filename string) {
	f, _ := os.Create(filename)
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	_ = writer.Write([]string{"path", "latitude", "longitude"})
	_ = writer.Write([]string{"file1.jpg", "51.5074", "0.1278"})
	_ = writer.Write([]string{"file2.jpg", "40.7128", "74.0060"})
}

// Helper function to create test images
func createTestImages(directory string) {
	_ = os.MkdirAll(directory, os.ModePerm)

	// Create dummy image files
	imageData := []byte{0x00, 0x01, 0x02, 0x03}
	_ = ioutil.WriteFile(filepath.Join(directory, "file1.jpg"), imageData, 0644)
	_ = ioutil.WriteFile(filepath.Join(directory, "file2.jpg"), imageData, 0644)
}

// Helper function to verify the generated CSV file
func verifyCSVFile(t *testing.T, filename string) {
	f, err := os.Open(filename)
	if err != nil {
		t.Errorf("Failed to open CSV file: %v", err)
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)

	// Verify column headers
	headers, err := reader.Read()
	if err != nil {
		t.Errorf("Failed to read CSV headers: %v", err)
		return
	}
	expectedHeaders := []string{"path", "latitude", "longitude"}
	if !compareStringSlices(headers, expectedHeaders) {
		t.Errorf("CSV headers mismatch. Expected: %v, Got: %v", expectedHeaders, headers)
	}

	// Verify data rows
	rows, err := reader.ReadAll()
	if err != nil {
		t.Errorf("Failed to read CSV rows: %v", err)
		return
	}
	expectedRows := [][]string{
		{"file1.jpg", "40.7128", "74.0060"},
		{"file2.jpg", "40.7128", "74.0060"},
		{"file3.jpg", "40.7128", "74.0060"},
		{"file4.jpg", "40.7128", "74.0060"},
		// {"file5.jpg", "40.7128", "74.0060"},
		// {"file6.jpg", "40.7128", "74.0060"},
	}
	if len(rows) != len(expectedRows) {
		t.Errorf("CSV rows count mismatch. Expected: %d, Got: %d", len(expectedRows), len(rows))
		return
	}
}

func compareStringSlices(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}
	return true
}

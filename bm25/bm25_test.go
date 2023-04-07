package bm25

import (
	"math"
	"os"
	"testing"
)

type TestData struct {
	Name  string
	Value int
}

func TestCompressAndWriteGZipFile(t *testing.T) { // Prepare test data
	filOpsImpl := FileOpsImpl{}

	data := TestData{
		Name:  "Test",
		Value: 42,
	}

	// Define file name and directory name
	fileName := "testfile.gz"
	dirName := "testdir"

	// Create a temporary directory
	if err := filOpsImpl.MkdirAll(dirName, 0755); err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}

	// Call the CompressAndWriteGzipFile function
	if err := filOpsImpl.CompressAndWriteGzipFile(fileName, data, dirName); err != nil {
		t.Fatalf("Failed to compress and write gzip file: %v", err)
	}

	// Check if the file exists
	filePath := dirName + "/" + fileName
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("File %s does not exist", filePath)
	}

	// Clean up: remove the file and the temporary directory
	if err := os.Remove(filePath); err != nil {
		t.Fatalf("Failed to remove temporary file: %v", err)
	}

	if err := os.Remove(dirName); err != nil {
		t.Fatalf("Failed to remove temporary directory: %v", err)
	}

	fileOpsNoImpl := FileOpsNoOp{}

	if err := fileOpsNoImpl.MkdirAll(dirName, 0755); err != nil {
		t.Fatalf("Error should be nil as this is a mocking DI: %v", err)
	}
	fileOpsNoImpl.CompressAndWriteGzipFile(fileName, data, dirName)
	// Check if the file exists
	filePath = dirName + "/" + fileName
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatalf("File %s does not exist", filePath)
	}

	// Clean up: remove the file and the temporary directory
	if err := os.Remove(filePath); err == nil {
		t.Fatalf("Expected error when removing non-existent temporary file: %v", err)
	}

	if err := os.Remove(dirName); err == nil {
		t.Fatalf("Expected error when removing temporary directory: %v", err)
	}

}

func TestFilterResults(t *testing.T) {
	results := []ResultsMap{
		{
			Name: "test",
			Path: "test",
			TF:   1,
		}, {
			Name: "test2",
			Path: "test2",
			TF:   0,
		},
	}

	filtered := FilterResults(results, IsGreaterThanZero)

	if len(filtered) != 1 {
		t.Errorf("FilterResults().len(filtered) == %d, want 1", len(filtered))
	}

	filtered = FilterResults(results, func(x float32) bool {
		return x > 1
	})

	if len(filtered) != 0 {
		t.Errorf("FilterResults().len(filtered) == %d, want 0", len(filtered))
	}
}

func TestResetResultsMap(t *testing.T) {

	results := []ResultsMap{
		{
			Name: "test",
			Path: "test",
			TF:   1,
		},
	}

	ResetResultsMap(results)

	if len(results) != 1 {
		t.Errorf("ResetResultsMap().len(results) == 0, want non-zero")
	}

}

func TestCalculateBm25(t *testing.T) {
	model := NewEmptyModel()

	LoadCachedGobToModel("../test-data/javascript.info", model)

	result, count := CalculateBm25(model, "javascript")

	if count < 1 {
		t.Errorf("CalculateBm25().count == %d, want greater than 1", count)
	}

	if len(result) == 0 {
		t.Errorf("CalculateBm25().len(result) == 0, want non-zero")
	}

}

func TestNewEmptyModel(t *testing.T) {
	model := NewEmptyModel()

	if model == nil {
		t.Errorf("NewEmptyModel() == nil, want non-nil")
	} else {
		if model.TFPD == nil {
			t.Errorf("NewEmptyModel().TFPD == nil, want non-nil")
		}

		if model.DF == nil {
			t.Errorf("NewEmptyModel().DF == nil, want non-nil")
		}

		if model.UrlFiles == nil {
			t.Errorf("NewEmptyModel().UrlFiles == nil, want non-nil")
		}

		if model.ModelLock == nil {
			t.Errorf("NewEmptyModel().ModelLock == nil, want non-nil")
		}
	}
}

func TestResetModel(t *testing.T) {
	model := NewEmptyModel()

	model.Name = "test"
	model.DocCount = 1
	model.TermCount = 1
	model.DirLength = 1
	model.DA = 1
	model.IsComplete = true

	model.TFPD["test"] = DocData{
		TermCount: 1,
		Terms:     TermFreq{"test": 1},
	}

	model.DF["test"] = 1
	model.UrlFiles["test"] = "test"

	ResetModel(model)

	if model.Name != "" {
		t.Errorf("ResetModel().Name == %s, want empty string", model.Name)
	}

	if model.DocCount != 0 {
		t.Errorf("ResetModel().DocCount == %d, want 0", model.DocCount)
	}

	if model.TermCount != 0 {
		t.Errorf("ResetModel().TermCount == %d, want 0", model.TermCount)
	}

	if model.DirLength != 0 {
		t.Errorf("ResetModel().DirLength == %f, want 0", model.DirLength)
	}

	if model.DA != 0 {
		t.Errorf("ResetModel().DA == %f, want 0", model.DA)
	}

	if model.IsComplete != false {
		t.Errorf("ResetModel().IsComplete == %t, want false", model.IsComplete)
	}

	if len(model.TFPD) != 0 {
		t.Errorf("ResetModel().TFPD == %d, want 0", len(model.TFPD))
	}

	if len(model.DF) != 0 {
		t.Errorf("ResetModel().DF == %d, want 0", len(model.DF))
	}

	if len(model.UrlFiles) != 0 {
		t.Errorf("ResetModel().UrlFiles == %d, want 0", len(model.UrlFiles))
	}

}

func TestReadUrlFiles(t *testing.T) {
	model := NewEmptyModel()

	readUrlFiles("../test-data/javascript.info", "url-files.gz", model, false)

	if len(model.UrlFiles) == 0 {
		t.Errorf("ReadUrlFiles() == 0, want non-zero")
	}

}

func TestReadCompressedFilesToModel(t *testing.T) {
	model := NewEmptyModel()

	readCompressedFilesToModel("../test-data/javascript.info", "indexed-data.gz", model)

	if len(model.TFPD) == 0 {
		t.Errorf("ReadCompressedFilesToModel() == 0, want non-zero")
	}

	if len(model.DF) == 0 {
		t.Errorf("ReadCompressedFilesToModel() == 0, want non-zero")
	}

	if model.DocCount == 0 {
		t.Errorf("ReadCompressedFilesToModel() == 0, want non-zero")
	}

}

func TestLoadCachedGobToModel(t *testing.T) {
	model := NewEmptyModel()

	LoadCachedGobToModel("../test-data/javascript.info", model)

	if len(model.TFPD) == 0 {
		t.Errorf("LoadCachedGobToModel() == 0, want non-zero")
	}

	if len(model.DF) == 0 {
		t.Errorf("LoadCachedGobToModel() == 0, want non-zero")
	}

	if model.DocCount == 0 {
		t.Errorf("LoadCachedGobToModel() == 0, want non-zero")
	}

	if len(model.UrlFiles) == 0 {
		t.Errorf("ReadUrlFiles() == 0, want non-zero")
	}
}

func TestComputeTF(t *testing.T) {
	testCases := []struct {
		name          string
		term          string
		totalTerms    int
		termFreq      TermFreq
		avgDocLength  float32
		expectedValue float32
	}{
		{
			name:          "Basic test",
			term:          "apple",
			totalTerms:    10,
			termFreq:      TermFreq{"apple": 3, "orange": 2, "banana": 5},
			avgDocLength:  5,
			expectedValue: 1.294118, // calculated manually
		},
		// Add more test cases here
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ComputeTF(tc.term, tc.totalTerms, tc.termFreq, tc.avgDocLength)
			if math.Abs(float64(result-tc.expectedValue)) > 1e-5 {
				t.Errorf("Expected: %f, got: %f", tc.expectedValue, result)
			}
		})
	}
}

func TestComputeIDF(t *testing.T) {
	testCases := []struct {
		name          string
		term          string
		totalDocs     int
		docFreq       DocFreq
		expectedValue float32
	}{
		{
			name:          "Basic test",
			totalDocs:     1000,
			docFreq:       DocFreq{"apple": 100, "orange": 50, "banana": 200},
			term:          "apple",
			expectedValue: 0.952318, // calculated manually
		},

		// Add more test cases here
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ComputeIDF(tc.term, tc.totalDocs, tc.docFreq)
			if math.Abs(float64(result-tc.expectedValue)) > 1e-5 {
				t.Errorf("Expected: %f, got: %f", tc.expectedValue, result)
			}

		})
	}
}

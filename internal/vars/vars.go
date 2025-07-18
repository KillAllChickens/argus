package vars

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var Version string = "v0.1.0"

// Argument var
var (
	Verbose bool
	AI      bool
	Threads int
	Silent  bool
)

// IO vars
var (
	ConfigDir             string
	ConfigJSONLocation    string
	ConfigSourcesLocation string
	PromptHTMLCheckFP     string
	// output vars
	OutputFolder string
	OutputTypes  []string
)

// Config variables
var (
	GeminiAPIKey string
)

// Scanner vars
var (
	Usernames  []string
	UserAgents []string
	// Options
	Proxy string
	// Tor   bool
)

// result vars
var (
	FoundSites map[string]map[string]string = make(map[string]map[string]string)
	FoundPFPs  map[string]map[string]string = make(map[string]map[string]string)
	// AI result vars
	AISiteSummaries map[string]string
	AITotalSummary  string
)

func InitConfVars() {
	var json map[string]any
	_, err := LoadAndStringifyJSON(ConfigJSONLocation, &json)
	if err != nil {
		os.Exit(1)
	}

	keys, ok := json["keys"].(map[string]any)
	if !ok {
		keys = make(map[string]any)
		json["keys"] = keys
	}

	GeminiAPIKey = keys["gemini"].(string)

	HTMLCheckFilePath, err := getFilePath("html_check.txt")
	if err != nil {
		os.Exit(1)
	}
	PromptHTMLCheckFP, err = getFileContent(HTMLCheckFilePath)
	if err != nil {
		os.Exit(1)
	}
}

func LoadAndStringifyJSON(path string, v any) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return "", err
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func getFilePath(name string) (string, error) {
	var FilePath string
	err := filepath.WalkDir(ConfigDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == name {
			FilePath = path
			return filepath.SkipDir // stop walking once found
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return FilePath, nil
}

func getFileContent(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

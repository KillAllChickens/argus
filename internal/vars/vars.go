package vars

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/KillAllChickens/argus/internal/printer"
)

var Version string = "v0.1.0"

// Deep Scan Resutl struct
type DeepScanResult struct {
	Description       *string   `json:"description,omitempty"`
	LinkedSocials     *[]string `json:"linked_socials,omitempty"`
	PublicPostCount   *int      `json:"public_post_count,omitempty"`
	FollowerCount     *int      `json:"follower_count,omitempty"`
	FollowingCount    *int      `json:"following_count,omitempty"`
	ProfilePictureURL *string   `json:"profile_picture_url,omitempty"`
	RealName *string   `json:"real_name,omitempty"`

}

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
	Proxies []string
	// Tor   bool
)

// result vars
var (
	FoundSites map[string]map[string]string = make(map[string]map[string]string)
	FoundPFPs  map[string]map[string]string = make(map[string]map[string]string)
	// AI result vars
	AISiteSummaries map[string]string
	AITotalSummary  string

	// Deep Scan related ones
	DeepScanEnabled bool
	DeepScanConfig  *map[string]DeepScanDomain
	DeepScanResults map[string]map[string]DeepScanResult = make(map[string]map[string]DeepScanResult)
)

// GEMINI

type DeepScanDomain struct {
	Targets []DeepScanTarget `json:"targets"`
}

type DeepScanTarget struct {
	Name     string           `json:"name"`
	Selector string           `json:"selector"`
	Actions  []DeepScanAction `json:"actions"`
}

type DeepScanAction struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

//

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

	deepScanConfigLocation, err := getFilePath("deepscan.json")
	if err == nil {
		_, err = LoadAndStringifyJSON(deepScanConfigLocation, &DeepScanConfig)
		if err != nil {
			// Handle error, maybe log it and disable deep scanning
			printer.Error("Could not import deepscan.json, continuing without deep scanning")
			DeepScanEnabled = false
		} else {
			DeepScanEnabled = true
		}
	} else {
		DeepScanEnabled = false
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

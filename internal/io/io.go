package io

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/KillAllChickens/argus/internal/helpers"
	"github.com/KillAllChickens/argus/internal/shared"
	"github.com/KillAllChickens/argus/internal/vars"
)

// vars.vars.ConfigJSONLocation

func InitPaths(CustomConfigPath string) {
	var configDir string
	var err error

	if CustomConfigPath != "" {
		vars.ConfigJSONLocation = CustomConfigPath
		configDir = filepath.Dir(CustomConfigPath)
	} else {
		configDir, err = GetConfigPath()
		helpers.HandleErr(err)

		ensureDefaultConfigExists(configDir)

		vars.ConfigJSONLocation = filepath.Join(configDir, "config.json")
	}

	vars.ConfigDir = configDir
	vars.ConfigSourcesLocation = filepath.Join(configDir, "sources.txt")
	shared.ArtworkFile = GetConfigFile("artworks.txt")
	UserAgenDir, err := GetFilePath("UserAgents.txt")
	helpers.HandleErr(err)
	UserAgents, err := NewlineSeperatedFileToArray(UserAgenDir)
	helpers.HandleErr(err)
	vars.UserAgents = UserAgents
}

func ensureDefaultConfigExists(configDir string) {
	pathExists, err := helpers.PathExists(configDir)
	helpers.HandleErr(err)

	if !pathExists {
		err := os.MkdirAll(configDir, 0755)
		helpers.HandleErr(err)
	}

	fileExists, _ := helpers.PathExists(filepath.Join(configDir, "config.json"))
	if !fileExists {
		err = CopyMissingConfigDir("./config", configDir)
		helpers.HandleErr(err)

		helpers.V("Copied default config files to " + configDir)
	}
}

func GetConfigPath() (string, error) {
	var configDir string

	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA not set")
		}
		configDir = filepath.Join(appData, "argus")
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, ".config", "argus")
	}

	// Ensure the directory exists
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", err
	}

	return configDir, nil
}

func FindNonGlobalConfigJSON(root string) (string, error) {
	var found string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err // failed to access this path
		}
		if !d.IsDir() && d.Name() == "config.json" {
			found = path
			return filepath.SkipDir // stop searching once we found one
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("config.json not found")
	}
	return found, nil
}

func CopyMissingConfigDir(srcDir, destDir string) error {
	return filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// Check if file exists
		_, err = os.Stat(destPath)
		if err == nil {
			// File already exists, skip it
			return nil
		}
		if !os.IsNotExist(err) {
			return err
		}

		// Copy the file
		input, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		err = os.WriteFile(destPath, input, 0644)
		if err != nil {
			return err
		}

		return nil
	})
}

func GetFilePath(name string) (string, error) {
	var FilePath string
	err := filepath.WalkDir(vars.ConfigDir, func(path string, d os.DirEntry, err error) error {
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

func GetSources() ([]string, error) {
	var sources []string
	var sourcesFilePath string

	// Find sources.txt in config dir
	err := filepath.WalkDir(vars.ConfigDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == "sources.txt" {
			sourcesFilePath = path
			return filepath.SkipDir // stop walking once found
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if sourcesFilePath == "" {
		return nil, fmt.Errorf("sources.txt not found")
	}

	// Read the file
	data, err := os.ReadFile(sourcesFilePath)
	if err != nil {
		return nil, err
	}

	// Split by newline and trim
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			sources = append(sources, line)
		}
	}

	return sources, nil
}

func ExportJSONConfig(newContent map[string]any) error {
	configPath := vars.ConfigJSONLocation

	newJsonData, err := json.MarshalIndent(newContent, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(configPath, newJsonData, 0644)
	if err != nil {
		return err
	}
	return nil
}

func NewlineSeperatedFileToArray(filePath string) ([]string, error) {
	var list []string

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Split by newline and trim
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Remove inline comment
		if idx := strings.Index(line, "#"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}
		// printer.Info("%s", line)
		if line != "" {
			list = append(list, line)
		}
	}

	return list, nil
}

func GetRandomUserAgent() string {
	// userAgentFilePath, err := GetFilePath("UserAgents.txt")
	// if err != nil {
	// 	return "Argus-Panoptes/0.0.1"
	// }
	// if userAgentFilePath == "" {
	// 	return "Argus-Panoptes/0.0.1"
	// }
	userAgents := vars.UserAgents
	return userAgents[rand.Intn(len(userAgents))]
}

func GetConfigFile(filename string) string {
	FilePath, err := GetFilePath(filename)
	helpers.HandleErr(err)
	f, err := os.Open(FilePath)
	helpers.HandleErr(err)

	defer func(){_ = f.Close()}()

	r := bufio.NewReader(f)

	data, err := io.ReadAll(r)
	helpers.HandleErr(err)

	return string(data)
}

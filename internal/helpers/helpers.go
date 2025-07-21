package helpers

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/KillAllChickens/argus/internal/colors"
	"github.com/KillAllChickens/argus/internal/printer"
	"github.com/KillAllChickens/argus/internal/vars"
)

func HandleErr(err error) {
	if err != nil {

		fmt.Fprintf(os.Stderr, "%s[!] ERROR: %v%s\n", colors.FgRed, err, colors.Reset)
		os.Exit(1)
	}
}

func V(format string, a ...any) {
	if vars.Verbose {
		printer.Info(format, a...)
		// fmt.Printf("DEBUG: format=%q, args=%v\n", format, a) // Debugging statement
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

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil // Path exists, no error
	}
	if os.IsNotExist(err) {
		return false, nil // Path does not exist
	}
	return false, err // Other error occurred
}

func ParseShorthandInt(s string) (int, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	multiplier := 1.0

	switch {
	case strings.HasSuffix(s, "K"):
		multiplier = 1_000
		s = strings.TrimSuffix(s, "K")
	case strings.HasSuffix(s, "M"):
		multiplier = 1_000_000
		s = strings.TrimSuffix(s, "M")
	case strings.HasSuffix(s, "B"):
		multiplier = 1_000_000_000
		s = strings.TrimSuffix(s, "B")
	}

	num, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}

	return int(math.Round(num * multiplier)), nil
}

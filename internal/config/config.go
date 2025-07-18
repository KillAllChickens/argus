package config

import (
	"bufio"
	"fmt"
	"github.com/KillAllChickens/argus/internal/colors"
	"github.com/KillAllChickens/argus/internal/helpers"
	"github.com/KillAllChickens/argus/internal/vars"
	"os"
	"strings"

	"github.com/KillAllChickens/argus/internal/io"
)

func RunConfig() {
	// io.InitPaths(CustomConfigPath string)
	helpers.V("Using config.json from: %s", vars.ConfigJSONLocation)

	var json map[string]any
	str, err := helpers.LoadAndStringifyJSON(vars.ConfigJSONLocation, &json)
	helpers.HandleErr(err)
	helpers.V("Current config: " + str)

	RunInputs(json)
}

func RunInputs(json map[string]any) {
	question := fmt.Sprintf("%s[?]%s Enter your Gemini API key(if you plan on using --ai, leave blank for no AI support)\n> ", colors.FgYellow, colors.Reset)
	fmt.Print(question)

	reader := bufio.NewReader(os.Stdin)
	GeminiKey, _ := reader.ReadString('\n')
	GeminiKey = strings.TrimSpace(GeminiKey)
	vars.GeminiAPIKey = GeminiKey

	keys, ok := json["keys"].(map[string]any)
	if !ok {
		keys = make(map[string]any)
		json["keys"] = keys
	}

	keys["gemini"] = vars.GeminiAPIKey

	helpers.V("Set gemini key to %s", vars.GeminiAPIKey)

	_ = io.ExportJSONConfig(json)

	helpers.V("Exported config to %s", vars.ConfigJSONLocation)

}

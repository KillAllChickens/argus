package ai

import (
	"context"
	"strings"
	"time"

	"github.com/KillAllChickens/argus/internal/helpers"
	"github.com/KillAllChickens/argus/internal/printer"
	"github.com/KillAllChickens/argus/internal/shared"
	"github.com/KillAllChickens/argus/internal/vars"

	"google.golang.org/genai"
)

var Limiter = NewTokenRateLimiter(1_000_000, 900_000)

func AIResponseWithRateLimit(system_prompt string, prompt string) string {
	if !vars.AI {
		return "true" // Include everything, including false positives
	}

	Limiter.waitIfNearLimit()

	ctx := context.Background()
	// client, err := genai.NewClient(ctx, option.WithAPIKey(vars.GeminiAPIKey))
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: vars.GeminiAPIKey,
	})
	helpers.HandleErr(err)
	// defer client.Close()

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(system_prompt, genai.RoleUser),
	}

	parts := genai.Text(prompt)

	resp, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.0-flash-lite",
		parts,
		config,
	)
	// helpers.HandleErr(err)
	if err != nil {
		if strings.Contains(err.Error(), "GenerateContentInputTokensPerModelPerMinute") || strings.Contains(err.Error(), "The model is overloaded") {
			shared.Bar.Clear()
			printer.Info("Hit AI quota limit for this minute, sleeping for 30 seconds and trying again.")
			time.Sleep(30 * time.Second)
			return AIResponse(system_prompt, prompt)
		}
		if strings.Contains(err.Error(), "GenerateContentInputTokensPerModelPerDay") {
			shared.Bar.Clear()
			printer.Info("Hit AI quota limit for today, continuing without AI.")
			vars.AI = false
			return AIResponse(system_prompt, prompt)
		}
	}
	// if err.Error != nil { // Happens when HandleErr returns without exitting the program
	// 	return "true"
	// }

	if resp != nil && resp.UsageMetadata != nil {
		Limiter.recordUsage(int(resp.UsageMetadata.TotalTokenCount))
	}

	return resp.Text()

}

func AIResponse(system_prompt string, prompt string) string {
	return AIResponseWithRateLimit(system_prompt, prompt)
}

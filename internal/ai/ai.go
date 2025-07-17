package ai

import (
	"context"
	"strings"
	"time"

	"argus/internal/helpers" // Assuming these are your local packages
	"argus/internal/printer"
	"argus/internal/shared"
	"argus/internal/vars"

	// Correct import for APIKey
	"google.golang.org/genai"
)

// Global instance of our rate limiter.
// It will be shared by all calls to AIResponseWithRateLimit.
var Limiter = NewTokenRateLimiter(1_000_000, 900_000)

// AIResponseWithRateLimit is the new function that wraps the original logic
// with our rate limiting behavior.
func AIResponseWithRateLimit(system_prompt string, prompt string) string {
	if !vars.AI {
		return "true" // Include everything, including false positives
	}

	// 1. Before doing anything, check if we need to wait.
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

	// 3. After a successful call, record the token usage.
	// The response contains metadata with the exact token count.
	if resp != nil && resp.UsageMetadata != nil {
		Limiter.recordUsage(int(resp.UsageMetadata.TotalTokenCount))
	}

	return resp.Text()

}

func AIResponse(system_prompt string, prompt string) string {
	return AIResponseWithRateLimit(system_prompt, prompt)
}

// package ai

// import (
// 	"context"

// 	"argus/internal/helpers"
// 	"argus/internal/vars"

// 	"google.golang.org/genai"
// )

// func AIResponse(system_prompt string, prompt string) string {
// 	if !vars.AI {
// 		return "false"
// 	}
// 	ctx := context.Background()

// 	client, err := genai.NewClient(ctx, &genai.ClientConfig{
// 		APIKey: vars.GeminiAPIKey,
// 	})

// 	helpers.HandleErr(err)

// 	config := &genai.GenerateContentConfig{
// 		SystemInstruction: genai.NewContentFromText(system_prompt, genai.RoleUser),
// 	}

// 	// result, err := client.GenerativeModel("gemini-1.5-flash").GenerateContent(
// 	// 	ctx,
// 	// 	genai.Text("Explain how AI works in a few words"), // Replace with 'parts' if using system_prompt/prompt
// 	// )
// 	result, err := client.Models.GenerateContent(
// 		ctx,
// 		"gemini-2.0-flash-lite",
// 		genai.Text(prompt),
// 		config,
// 	)

// 	helpers.HandleErr(err)

// 	return result.Text()
// }

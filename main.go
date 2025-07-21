package main

import (
	"context"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/KillAllChickens/argus/internal/config"
	"github.com/KillAllChickens/argus/internal/helpers"
	"github.com/KillAllChickens/argus/internal/io"
	"github.com/KillAllChickens/argus/internal/printer"
	"github.com/KillAllChickens/argus/internal/scanner"
	"github.com/KillAllChickens/argus/internal/vars"
	"github.com/skratchdot/open-golang/open"
)

var usernames []string

// type author struct {
// 	Name string
// 	Email string
// }

func main() {
	cmd := &cli.Command{
		// Base config
		EnableShellCompletion: true,
		Suggest:               true,
		Name:                  "argus",
		Version:               vars.Version,
		Copyright:             "(c) 2025 KillAllChickens (KAC)",
		Usage:                 "The all-seeing OSINT username search tool",
		// Authors: []any{
		// 	author{
		// 		Name: "Vance Perry",
		// 		Email: "vance@killallchickens.org",
		// 	},
		// },
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "verbose", Aliases: []string{"V"}},
			&cli.StringFlag{
				Name:    "config-path",
				Aliases: []string{},
				Value:   "",
				Usage:   "The path to a custom config.json, leave blank for default",
				Hidden:  true,
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			vars.Verbose = cmd.Bool("verbose")
			cmd.String("config-path")
			// vars.AI = cmd.Bool("ai")
			return nil, nil
		},
		Commands: []*cli.Command{
			{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Configure Argus to use AI.",
				// Commands: []*cli.Command{},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					io.InitPaths(cmd.String("config-path"))
					config.RunConfig()

					return nil
				},
			},
			{
				Name:          "scan",
				Usage:         "Scan username(s).",
				Aliases:       []string{"s"},
				ShellComplete: func(ctx context.Context, cmd *cli.Command) {}, // Do nothing for completion(allowing autocomplete files). Odd bug with urfave/cli
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "threads",
						Usage:   "Amount of concurrent requests",
						Aliases: []string{"t"},
						Value:   25,
					},
					&cli.BoolFlag{
						Name:  "ai",
						Usage: "Use AI to eliminate false positives. (Increases scan time)",
					},
					&cli.StringFlag{
						Name:        "username-list",
						Aliases:     []string{"u"},
						DefaultText: "",
						Usage:       "Get usernames to scan, one per line",
						TakesFile:   true,
					},
					&cli.StringFlag{
						Name:        "output",
						Aliases:     []string{"o"},
						DefaultText: "",
						Usage:       "The directory to output to, defaults to ./results/. if you don't specify a specific type, it will output all types",
					},

					&cli.StringFlag{
						Name:        "proxy",
						Aliases:     []string{"p"},
						DefaultText: "",
						Usage:       "Proxy to use for scanning (e.g., http://proxyserver:8888 or socks5://user:pass@proxyserver:port)",
					},
					&cli.StringFlag{
						Name:        "proxy-list",
						Aliases:     []string{"pl"},
						DefaultText: "",
						Usage:       "List of proxied to use, one per line.",
					},
					&cli.BoolFlag{Name: "tor", Usage: "Use Tor for scanning"},

					&cli.BoolFlag{Name: "silent", Aliases: []string{"s"}, Usage: "Disable \"Scan Complete\" notifications.", Destination: &vars.Silent},

					&cli.BoolFlag{Name: "deep-scan", Aliases: []string{"ds"}, Usage: "Run a Deep Scan, will try to collect more information", Destination: &vars.DeepScanEnabled},

					// Output types
					&cli.BoolFlag{Name: "html", Usage: "Output as HTML"},
					&cli.BoolFlag{Name: "pdf", Usage: "Output as PDF"},
					&cli.BoolFlag{Name: "json", Usage: "Output as JSON"},
					&cli.BoolFlag{Name: "text", Aliases: []string{"txt"}, Usage: "Output as Text"},
				},
				Arguments: []cli.Argument{
					&cli.StringArgs{
						Name: "usernames",
						// UsageText:   "usernames",
						Destination: &usernames,
						Max:         -1,
						Min:         0,
					},
				},
				Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {

					vars.AI = cmd.Bool("ai")
					if cmd.String("output") == "" {
						vars.OutputFolder = "./results/"
					} else {
						vars.OutputFolder = cmd.String("output")
					}

					// Output types
					if cmd.Bool("html") {
						vars.OutputTypes = append(vars.OutputTypes, "html")
					}
					if cmd.Bool("pdf") {
						vars.OutputTypes = append(vars.OutputTypes, "pdf")
					}
					if cmd.Bool("json") {
						vars.OutputTypes = append(vars.OutputTypes, "json")
					}
					if cmd.Bool("text") {
						vars.OutputTypes = append(vars.OutputTypes, "text")
					}

					if cmd.String("proxy") != "" && cmd.String("proxy-list") != "" {
						printer.Error("Cannot use --proxy/-p with --proxy-list/-pl. You must choose one or the other.")
						os.Exit(1)
					}

					if cmd.String("proxy") != "" && cmd.Bool("tor") {
						printer.Error("Cannot use --proxy/-p with --tor. You must choose one or the other.")
						os.Exit(1)
					}

					if cmd.String("proxy") != "" {
						vars.Proxies = append(vars.Proxies, cmd.String("proxy"))
					}

					if cmd.Bool("tor") {
						vars.Proxies = append(vars.Proxies, "socks5://127.0.0.1:9050")
					}

					if cmd.String("proxy-list") != "" {
						vars.Proxies, _ = io.NewlineSeperatedFileToArray(cmd.String("proxy-list"))
					}

					vars.Threads = cmd.Int("threads")

					return ctx, nil
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					var usernames_list []string
					if cmd.String("username-list") != "" {
						usernames_file, err := io.NewlineSeperatedFileToArray(cmd.String("username-list"))
						helpers.HandleErr(err)
						usernames_list = usernames_file
						if len(usernames) > 0 {
							usernames_list = append(usernames_list, usernames...)
						}
					} else {
						usernames_list = usernames
					}
					if len(usernames_list) == 0 {
						printer.Error("At least one username is required!")
						return cli.ShowSubcommandHelp(cmd)
					}
					original := usernames_list
					usernames_list = []string{}

					for _, username := range original {
						if strings.Contains(username, "{?}") {
							for _, replacement := range []string{"", "-", "_"} {
								result := strings.ReplaceAll(username, "{?}", replacement)
								usernames_list = append(usernames_list, result)
							}
						} else {
							usernames_list = append(usernames_list, username)
						}
					}

					vars.Usernames = usernames_list

					scanner.Init(cmd.String("config-path"))

					if vars.AI {
						if vars.GeminiAPIKey == "" {
							printer.Error("You must configure Argus with a Google Gemini API key in order to use --ai. Run '%s config'", cmd.Root().Name)
							_ = cli.ShowAppHelp(cmd) // Use _ to ignore the error
							return nil
						}
					}
					scanner.StartScan(vars.Usernames)

					return nil
				},
			},
			{
				Name: "config-dir",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					io.InitPaths(cmd.String("config-path"))
					err := open.Run(vars.ConfigDir)
					helpers.HandleErr(err)
					return nil
				},
			},
		},
	}

	isScanHelp := false
	for _, arg := range os.Args { // custom solution for a "bug" in urfave/cli Without an obvious solution
		if arg == "scan" || arg == "s" {
			for _, helpArg := range os.Args {
				if helpArg == "--help" || helpArg == "-h" {
					isScanHelp = true
					break
				}
			}
			break
		}
	}

	if isScanHelp {
		// Manually find the subcommand and show its help text.
		for _, c := range cmd.Commands {
			if c.Name == "scan" {
				c.Root().Writer = os.Stdout
				// break
				cli.ShowSubcommandHelpAndExit(c, 0)
			}
		}
	}

	err := cmd.Run(context.Background(), os.Args)
	helpers.HandleErr(err)
}

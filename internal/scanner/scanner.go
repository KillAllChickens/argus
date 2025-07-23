package scanner

import (
	"crypto/rand"
	"fmt"
	"math/big"
	mrand "math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/KillAllChickens/argus/internal/ai"
	"github.com/KillAllChickens/argus/internal/colors"
	"github.com/KillAllChickens/argus/internal/helpers"
	"github.com/KillAllChickens/argus/internal/io"
	"github.com/KillAllChickens/argus/internal/output"
	"github.com/KillAllChickens/argus/internal/printer"
	"github.com/KillAllChickens/argus/internal/shared"
	"github.com/KillAllChickens/argus/internal/vars"

	"github.com/PuerkitoBio/goquery"
	"github.com/gen2brain/beeep"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/net/publicsuffix"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"resty.dev/v3"
)

var soft404Fingerprints = []string{
	"user not found",
	"profile not found",
	"could not find user",
	"this profile is not available",
	"user does not exist",
	"the user you are looking for does not exist",
	"account not found",
	"no user with that name",
	"username is not available",
	"this user does not have a profile",
	"page not found",
	"the page you were looking for",
	"couldn't find this page",
	"the requested url was not found on this server",
	"we can't find that page",
	"the resource cannot be found",
	"this page is not available",
	"the page you are looking for is not here",
	"there isn't a page here",
	"sorry, we can't find that page",
	"sorry, this page is not available",
	"oops, that page can't be found",
	"whoops, something went wrong",
	"uh oh, page not found",
	"it looks like nothing was found at this location",
	"the link you followed may be broken",
	"check the url for typos",
	// Basic checks
	"404 not found",
	"404 error",
	// To configure specific checks, use <CONFIG>/404checks.txt
}

var badRedirects []string // will be set based on <CONFIG>/BadRedirects.txt

func StartScan(usernames []string) {
	printer.AsciiArtwork()
	printer.Info("Starting Argus %s", vars.Version)
	init404Checks()
	initBadRedirects()
	if vars.AI {
		printer.Info("Running with Google Gemini capabilities")
	}

	client := resty.New()
	client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(20))
	client.SetTimeout(5 * time.Second)

	if len(vars.Proxies) == 1 {
		proxyTest := testProxy(vars.Proxies[0])
		if proxyTest { // Proxy works as expected
			client.SetProxy(vars.Proxies[0])
		} else {
			if vars.Proxies[0] == "socks5://127.0.0.1:9050" {
				printer.Info("Do you have Tor installed and set up?")
			}
			os.Exit(1)
		}
	} else if len(vars.Proxies) > 1 {
		if len(vars.Proxies) <= 10 {
			printer.Info("Testing %d proxies", len(vars.Proxies))
			for _, proxy := range vars.Proxies {
				proxyTest := testProxy(proxy)
				if proxyTest { // Proxy works as expected
					printer.Success("Proxy %s works!", proxy)
				} else {
					printer.Error("%s is an invalid proxy, removing from list and continuing.", proxy)
					remove(vars.Proxies, proxy)
					continue
				}
			}
		} else {
			printer.Info("Running with %d proxies", len(vars.Proxies))
		}
		randomProxy := vars.Proxies[mrand.Intn(len(vars.Proxies))]
		client.SetProxy(randomProxy)
	}

	defer func() { _ = client.Close() }()

	sources, err := io.GetSources()
	helpers.HandleErr(err)

	for i, username := range usernames {
		scanDesc := fmt.Sprintf("%s[%d/%d]%s Searching '"+username+"'", colors.FgGreen, i+1, len(usernames), colors.Reset)
		bar := progressbar.NewOptions(len(sources),
			progressbar.OptionSetWriter(os.Stdout),
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionSetDescription(scanDesc),
			progressbar.OptionShowElapsedTimeOnFinish(),
			progressbar.OptionShowCount(),
			progressbar.OptionClearOnFinish(),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        colors.FgGreen + "█" + colors.Reset,
				SaucerHead:    colors.FgGreen + "▒" + colors.Reset,
				SaucerPadding: " ",
				BarStart:      "[",
				BarEnd:        "]",
			}))
		shared.Bar = bar
		// bar.Describe(scanDesc)
		// bar := progressbar.Default(100, scanDesc)

		var wg sync.WaitGroup
		var mtx sync.Mutex

		jobs := make(chan string, len(sources))
		var numWorkers int
		if vars.AI {
			numWorkers = 10
		} else {
			numWorkers = vars.Threads
		}

		for w := 1; w <= numWorkers; w++ {
			go func(id int, jobs <-chan string, wg *sync.WaitGroup, u string) {
				for source := range jobs {
					FetchSource(client, u, source, bar, &mtx)
					wg.Done()
				}
			}(w, jobs, &wg, username)
		}

		for _, source := range sources {
			wg.Add(1)
			jobs <- source
		}
		close(jobs)
		wg.Wait()
		if i+1 != len(usernames) {
			printer.Info("Finished search on %s, starting %s", username, usernames[i+1])
		} else {
			printer.Info("Finished search on %s", username)
		}
	}
	CompleteScanning()
}

func init404Checks() {
	checkfilepath, err := io.GetFilePath("404checks.txt")
	helpers.HandleErr(err)
	checks, _ := io.NewlineSeperatedFileToArray(checkfilepath)
	soft404Fingerprints = append(soft404Fingerprints, checks...)
}

func initBadRedirects() {
	checkfilepath, err := io.GetFilePath("BadRedirects.txt")
	helpers.HandleErr(err)
	fileContent, _ := io.NewlineSeperatedFileToArray(checkfilepath)
	badRedirects = fileContent
	for i, badRedirect := range badRedirects {
		badRedirects[i] = normalizeURL(badRedirect)
	}
}

func Init(CustomConfigPath string) {
	io.InitPaths(CustomConfigPath)
	vars.InitConfVars()
}

// func FetchSource(username string, source string, bar *progressbar.ProgressBar) {
func FetchSource(client *resty.Client, username string, source string, bar *progressbar.ProgressBar, mtx *sync.Mutex) {
	// client := resty.New()

	// defer client.Close()

	// client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(5))
	// defer bar.Add(1)
	//
	if len(vars.Proxies) > 1 {
		randomProxy := vars.Proxies[mrand.Intn(len(vars.Proxies))]
		client.SetProxy(randomProxy)
	}

	defer func() { _ = bar.Add(1) }()

	parts := strings.Split(source, "|")
	URL := parts[len(parts)-1]
	URL = strings.ReplaceAll(URL, "{U}", username)

	reqURL := parts[0]
	reqURL = strings.ReplaceAll(reqURL, "{U}", username)

	client.SetRedirectPolicy(resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
		if vars.Verbose {
			mtx.Lock()
			_ = bar.Clear()
			printer.Error("Redirect: %s -> %s", via[len(via)-1].URL.String(), req.URL.String())
			mtx.Unlock()
		}
		for _, badRedirect := range badRedirects {
			if normalizeURL(req.URL.String()) == strings.ReplaceAll(badRedirect, "{U}", username) {
				if vars.Verbose {
					mtx.Lock()
					_ = bar.Clear()
					printer.Error("Bad Redirect: tried going to %s from %s", badRedirect, req.URL.String())
					mtx.Unlock()
				}
				return fmt.Errorf("bad redirect: tried going to %s from %s", badRedirect, req.URL.String())
			}
		}
		return nil
	}))

	// client.OnBeforeRedirect(func (req *resty.Request, via []*http.Request) error {
	// 	for _, badRedirect := range badRedirects {
	// 	if strings.Contains(badRedirect, req.URL) {

	// 	}
	// 	}
	// })

	res, err := client.R().
		SetHeader("User-Agent", io.GetRandomUserAgent()).
		Get(reqURL)
	// helpers.HandleErr(err)
	if err != nil {
		if vars.Verbose {
			mtx.Lock()
			_ = bar.Clear()
			printer.Error("Network error for %s: %v", reqURL, err)
			mtx.Unlock()
		}
		return
	}

	if res.IsError() {
		switch res.StatusCode() {
		case http.StatusNotFound, http.StatusGone:
			if vars.Verbose {
				mtx.Lock()
				_ = bar.Clear()
				printer.Error("'%s' not found in %s (Status: %d)", username, reqURL, res.StatusCode())
				mtx.Unlock()
			}
		default:
			if vars.Verbose {
				mtx.Lock()
				_ = bar.Clear()
				printer.Error("Received error status %d for '%s' at %s", res.StatusCode(), username, reqURL)
				mtx.Unlock()
			}
		}
		return
	}
	if res.IsSuccess() {
		body := res.String()
		bodyLower := strings.ToLower(body)
		usernameLower := strings.ToLower(username)

		if !strings.Contains(bodyLower, usernameLower) {
			if vars.Verbose {
				mtx.Lock()
				_ = bar.Clear()
				printer.Error("'%s' not found in %s (Soft 404 detected, username not in body)", username, URL)
				mtx.Unlock()
			}
			return
		}

		for _, fingerprint := range soft404Fingerprints {
			fingerprint = strings.ReplaceAll(fingerprint, "{U}", usernameLower)
			fingerprint = strings.ToLower(fingerprint)

			if strings.Contains(bodyLower, fingerprint) {
				if vars.Verbose {
					mtx.Lock()
					_ = bar.Clear()
					printer.Error("'%s' not found in %s (Soft 404)", username, URL)
					mtx.Unlock()
				}
				return
			}
		}

		// Last and final check, against a non-existent user
		nonExistantUsername, err := generateUsername(30)
		if err == nil {
			nonExistentUserURL := strings.ReplaceAll(reqURL, "{U}", nonExistantUsername)
			testBody := getBodyLower(client, nonExistentUserURL)
			testBody = strings.ReplaceAll(testBody, nonExistentUserURL, usernameLower)
			if testBody != "" && testBody == bodyLower {
				if vars.Verbose {
					mtx.Lock()
					_ = bar.Clear()
					printer.Error("'%s' not found in %s (Same as non-existent user)", username, URL)
					mtx.Unlock()
				}
				return
			}
		}

		prompt := strings.ReplaceAll(vars.PromptHTMLCheckFP, "{S}", URL)
		prompt = strings.ReplaceAll(prompt, "{U}", username)
		// mtx.Lock()
		verdict := strings.ToLower(ai.AIResponse(prompt, res.String()))
		// mtx.Unlock()
		verdict = strings.TrimSpace(verdict)
		if vars.Verbose && vars.AI {
			mtx.Lock()
			_ = bar.Clear()
			helpers.V("AI says '%s' for %s", verdict, URL)
			mtx.Unlock()
		}
		if verdict == "true" {
			mtx.Lock()
			_ = bar.Clear()
			printer.Success("FOUND: %s", URL)
			MainDomain, err := GetMainDomain(URL)
			if vars.FoundSites[username] == nil {
				vars.FoundSites[username] = make(map[string]string)
			}
			vars.FoundSites[username][MainDomain] = URL
			PFPUrl := ExtractPFP(body, URL)
			if PFPUrl != "" {
				if err != nil {
					mtx.Lock()
					_ = bar.Clear()
					printer.Error("Failed to get main domain for %s: %s", URL, err)
					mtx.Unlock()
					return
				}
				// printer.Success("Found PFP for %s: %s", MainDomain, PFPUrl)
				if vars.FoundPFPs[username] == nil {
					vars.FoundPFPs[username] = make(map[string]string)
				}
				vars.FoundPFPs[username][MainDomain] = PFPUrl
			}

			if vars.DeepScanEnabled {
				if domainConfig, ok := (*vars.DeepScanConfig)[MainDomain]; ok {
					deepScanResult := performDeepScan(res.String(), domainConfig)
					if vars.DeepScanResults[username] == nil {
						vars.DeepScanResults[username] = make(map[string]vars.DeepScanResult)
					}
					vars.DeepScanResults[username][MainDomain] = deepScanResult
				}
			}

			mtx.Unlock()

		}
	}
}

type selectorStrategy struct {
	selector  string
	attribute string
}

func ExtractPFP(body string, baseURL string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return ""
	}

	strategies := []selectorStrategy{
		{`meta[property="og:image"]`, "content"},
		{`meta[property="og:image:secure_url"]`, "content"},
		{`meta[name="twitter:image"]`, "content"},
		{`meta[name="twitter:image:src"]`, "content"},
		{`img[class*="avatar"]`, "src"},
		{`img[class*="profile"]`, "src"},
		{`img[id*="avatar"]`, "src"},
		{`img[id*="profile"]`, "src"},
		{`img[alt*="avatar"]`, "src"},
		{`img[alt*="profile"]`, "src"},
		{`article img[src]`, "src"},
		{`header img[src]`, "src"},
		{`link[rel="apple-touch-icon"]`, "href"},
		{`link[rel="icon"]`, "href"},
		{`link[rel="shortcut icon"]`, "href"},
	}

	for _, s := range strategies {
		selection := doc.Find(s.selector).First()
		if selection.Length() > 0 {
			imageURL, exists := selection.Attr(s.attribute)
			if exists && imageURL != "" {
				resolvedURL, err := resolveURL(baseURL, imageURL)
				if err == nil {
					return resolvedURL
				}
			}
		}
	}
	return ""
}

func resolveURL(base, image string) (string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("could not parse base URL: %w", err)
	}

	imageURL, err := url.Parse(image)
	if err != nil {
		return "", fmt.Errorf("could not parse image URL: %w", err)
	}

	return baseURL.ResolveReference(imageURL).String(), nil
}

func GetMainDomain(rawURL string) (string, error) {
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("could not parse url: %w", err)
	}

	hostname := parsedURL.Hostname()
	if hostname == "" {
		return "", fmt.Errorf("could not extract hostname from url")
	}

	mainDomain, err := publicsuffix.EffectiveTLDPlusOne(hostname)
	if err != nil {
		return "", fmt.Errorf("could not determine main domain for '%s': %w", hostname, err)
	}

	return mainDomain, nil
}

func CompleteScanning() {
	for _, username := range vars.Usernames {
		// printer.Success("Found %d sites for %s!", len(vars.FoundSites[username]), username)
		printer.Info("All sites for %s:", username)
		for n, site := range vars.FoundSites[username] {
			printer.Success("%-14s => %-45s", n, site)
			if deepScanData, ok := vars.DeepScanResults[username][n]; ok {
				if deepScanData.Description != nil {
					printer.Info("  Description: %s", *deepScanData.Description)
				}
				if deepScanData.FollowerCount != nil {
					printer.Info("  Followers: %d", *deepScanData.FollowerCount)
				}
				if deepScanData.FollowingCount != nil {
					printer.Info("  Following: %d", *deepScanData.FollowingCount)
				}
				if deepScanData.RealName != nil {
					printer.Info("  Real Name: %s", *deepScanData.RealName)
				}
				if len(deepScanData.NonDefinedActions) > 0 {
					for _, action := range deepScanData.NonDefinedActions {

						printer.Info("  %s: %s", action.Name, action.Value)
					}
				}
			}
		}
	}
	if len(vars.OutputTypes) == 0 {
		printer.Success("Scanning complete!")
		compExit()
	}

	if vars.OutputFolder != "" {
		err := os.MkdirAll(vars.OutputFolder, 0755)
		helpers.HandleErr(err)
		if len(vars.Usernames) > 1 {
			for _, username := range vars.Usernames {
				err := os.MkdirAll(filepath.Join(vars.OutputFolder, username), 0755)
				helpers.HandleErr(err)
			}
		}
	}
	if len(vars.OutputTypes) != 0 {
		printer.Info("Outputting results to %s", vars.OutputFolder)
		for _, outputType := range vars.OutputTypes {
			switch outputType {
			case "html":
				output.OutputHTML()
			case "json":
				output.OutputJSON()
			case "pdf":
				output.OutputPDF()
			case "text":
				output.OutputText()
			default:
				printer.Error("Unknown output type: %s", outputType)
			}
		}
	}
	printer.Success("Scanning complete!")
	compExit()
}

func getBodyLower(client *resty.Client, url string) string {

	// URL = strings.ReplaceAll(URL, "{U}", username)

	res, err := client.R().
		SetHeader("User-Agent", io.GetRandomUserAgent()).
		Get(url)
	if err != nil || res.IsError() {
		return ""
	}
	return strings.ToLower(res.String())
}

func generateUsername(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}

	return string(result), nil
}

func compExit() {
	if !vars.Silent {
		var resCount int
		for _, res := range vars.FoundSites {
			resCount += len(res)
		}
		beeep.AppName = "Argus"
		err := beeep.Alert("Argus scan complete!", fmt.Sprintf("Found %d results for %d username(s)", resCount, len(vars.Usernames)), "")
		helpers.HandleErr(err)
	}
	os.Exit(0)
}

func normalizeURL(fullURL string) string {
	u1, err := url.Parse(fullURL)
	if err != nil {
		return ""
	}

	host1 := strings.TrimPrefix(strings.ToLower(u1.Host), "www.")

	path1 := strings.TrimRight(u1.Path, "/")

	if path1 == "" {
		path1 = "/"
	}
	normalizedFull := host1 + path1
	if u1.RawQuery != "" {
		normalizedFull += "?" + u1.RawQuery
	}

	return normalizedFull
}

func testProxy(proxyAddr string) bool {
	client := resty.New()
	client.SetProxy(proxyAddr)
	client.SetTimeout(10 * time.Second)

	client.SetRedirectPolicy(resty.NoRedirectPolicy())
	resp, err := client.R().Get("http://ipinfo.io/ip")

	if err != nil {
		if strings.Contains(err.Error(), "proxyconnect") ||
			strings.Contains(err.Error(), "connection refused") ||
			strings.Contains(err.Error(), "timeout") {
			printer.Error("Proxy %s failed to connect or timed out: %v", proxyAddr, err)
		} else if strings.Contains(err.Error(), "protocol error") {
			printer.Error("Proxy %s had a protocol error. Is it the correct type (http/socks5)? %v", proxyAddr, err)
		} else {
			printer.Error("Failed to use proxy %s for request: %v", proxyAddr, err)
		}
		return false
	}

	if resp.StatusCode() != http.StatusOK {
		if resp.StatusCode() == http.StatusForbidden || resp.StatusCode() == http.StatusProxyAuthRequired {
			printer.Warning("Proxy %s might require authentication or is blocked.", proxyAddr)
		}
		return false
	}

	return true
}

func performDeepScan(body string, config vars.DeepScanDomain) vars.DeepScanResult {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return vars.DeepScanResult{}
	}

	result := vars.DeepScanResult{}

	for _, target := range config.Targets {
		selection := doc.Find(target.Selector)
		helpers.V("target.Selector: %s", target.Selector)
		text := strings.TrimSpace(selection.First().Text())

		// Apply actions
		for _, action := range target.Actions {
			if action.Type == "ignore_contains" && strings.Contains(text, action.Value) {
				text = ""
				break
			}
		}

		if text == "" {
			continue
		}

		// Map the extracted data to the DeepScanResult struct
		switch target.Name {
		case "description":
			result.Description = &text
		case "follower_count":
			if followerCount, err := helpers.ParseShorthandInt(strings.ReplaceAll(text, ",", "")); err == nil {
				result.FollowerCount = &followerCount
			}
		case "following_count":
			if followingCount, err := helpers.ParseShorthandInt(strings.ReplaceAll(text, ",", "")); err == nil {
				result.FollowingCount = &followingCount
			}
		case "real_name":
			result.RealName = &text
		default:
			caser := cases.Title(language.English)
			actionName := strings.ReplaceAll(target.Name, "_", " ")
			actionName = strings.TrimSpace(actionName)
			actionName = caser.String(actionName)
			result.NonDefinedActions = append(result.NonDefinedActions, vars.NonDefinedAction{Name: actionName, Value: text})
		}
	}

	return result
}

func remove[T comparable](l []T, item T) []T {
	out := make([]T, 0)
	for _, element := range l {
		if element != item {
			out = append(out, element)
		}
	}
	return out
}

package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/KillAllChickens/argus/internal/helpers"
	"github.com/KillAllChickens/argus/internal/io"
	"github.com/KillAllChickens/argus/internal/vars"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/jung-kurt/gofpdf"
)

type jsonSiteResult struct {
	URL      string               `json:"url"`
	DeepScan *vars.DeepScanResult `json:"deep_scan_results,omitempty"`
}

type outputJSONStruct struct {
	Username  string                    `json:"username"`
	Timestamp string                    `json:"timestamp"`
	Results   map[string]jsonSiteResult `json:"sites"`
}

// for pdf file
type color struct {
	r, g, b int
}

func OutputJSON() {
	for _, username := range vars.Usernames {
		data := outputJSONStruct{
			Username:  username,
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			Results:   make(map[string]jsonSiteResult),
		}

		for siteName, siteURL := range vars.FoundSites[username] {
			result := jsonSiteResult{
				URL: siteURL,
			}
			if deepScanData, ok := vars.DeepScanResults[username][siteName]; ok {
				result.DeepScan = &deepScanData
			}
			data.Results[siteName] = result
		}

		jsonData, err := json.MarshalIndent(data, "", "  ")
		helpers.HandleErr(err)

		saveResultFile("json", username, string(jsonData))
	}
}

func OutputHTML() {
	// tmpl, err := template.New("results")
	for _, username := range vars.Usernames {
		tpl := io.GetConfigFile("html_template.html")

		data := map[string]any{
			"Username":        username,
			"Sites":           vars.FoundSites[username],
			"PFPs":            vars.FoundPFPs[username],
			"DeepScanEnabled": vars.DeepScanEnabled,
			"DeepScans":       vars.DeepScanResults[username],
			"Timestamp":       time.Now().Format("2006-01-02 15:04:05"),
			"Version":         vars.Version,
		}

		funcMap := template.FuncMap{
			"index": func(m map[string]string, key string) string {
				return m[key]
			},
			"getDeepScan": func(m map[string]vars.DeepScanResult, key string) *vars.DeepScanResult {
				if val, ok := m[key]; ok {
					return &val
				}
				return nil
			},
		}

		var buf bytes.Buffer
		t := template.Must(template.New("test").Funcs(funcMap).Parse(tpl))
		if err := t.Execute(&buf, data); err != nil {
			helpers.HandleErr(err)
		}

		saveResultFile("html", username, buf.String())
	}
}

func OutputText() {
	header := "==================================================\n" +
		"           Argus Scan Results\n" +
		"==================================================\n" +
		"Username: {U}\n" +
		"Timestamp: {T}\n"
	for _, username := range vars.Usernames {
		fullText := strings.ReplaceAll(header, "{U}", username)
		fullText = strings.ReplaceAll(fullText, "{T}", time.Now().Format("2006-01-02 15:04:05"))
		fullText += "--------------------------------------------------\n"

		for siteName, siteURL := range vars.FoundSites[username] {
			fullText += fmt.Sprintf("[+] %-14s => %-45s\n", siteName, siteURL)

			// Check for and append deep scan results.
			if deepResult, ok := vars.DeepScanResults[username][siteName]; ok {
				// Use reflection to iterate through the fields of the DeepScanResult struct.
				val := reflect.ValueOf(deepResult)
				typ := val.Type()
				for i := 0; i < val.NumField(); i++ {
					field := val.Field(i)
					if field.IsNil() {
						continue // Skip empty/nil fields
					}

					// Get the field name from the json tag for cleaner output.
					fieldName := strings.Split(typ.Field(i).Tag.Get("json"), ",")[0]

					var fieldValue string
					// Handle different types of fields.
					switch f := field.Elem().Interface().(type) {
					case string:
						fieldValue = f
					case int:
						fieldValue = strconv.Itoa(f)
					case []string:
						fieldValue = strings.Join(f, ", ")
					case []vars.NonDefinedAction:
						var nonDefinedStrs []string
						for _, action := range f {
							nonDefinedStrs = append(nonDefinedStrs, fmt.Sprintf("%s: %s", action.Name, action.Value))
						}
						fieldValue = strings.Join(nonDefinedStrs, "; ")
					default:
						continue
					}

					if fieldValue != "" {
						fullText += fmt.Sprintf("  - %-18s: %s\n", strings.ReplaceAll(fieldName, "_", " "), fieldValue)
					}
				}
			}
		}
		fullText += "--------------------------------------------------\n"
		fullText += fmt.Sprintf("%d sites found for %s\n", len(vars.FoundSites[username]), username)
		saveResultFile("txt", username, fullText)
	}
}

func OutputPDF() {
	const (
		pageMargin   = 10.0
		pageWidth    = 210.0
		pageHeight   = 297.0
		cardPadding  = 5.0
		siteColWidth = 45.0
	)

	headerBgColor := color{226, 232, 240} // Light Slate Gray
	headerTextColor := color{40, 40, 40}
	// borderColor := color{203, 213, 225} // Slate
	primaryTextColor := color{23, 23, 23}
	secondaryTextColor := color{100, 116, 139} // Lighter Slate for URLs
	accentColor := color{37, 99, 235}          // Blue for links/headers

	for _, username := range vars.Usernames {
		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.SetMargins(pageMargin, pageMargin, pageMargin)
		pdf.AddPage()

		drawHeader := func() {
			pdf.SetFont("Arial", "B", 20)
			pdf.SetTextColor(headerTextColor.r, headerTextColor.g, headerTextColor.b)
			pdf.CellFormat(0, 15, "Argus Scan Results", "", 1, "C", false, 0, "")

			pdf.SetFont("Arial", "", 12)
			pdf.SetTextColor(secondaryTextColor.r, secondaryTextColor.g, secondaryTextColor.b)
			pdf.CellFormat(0, 10, "Username: "+username, "", 0, "L", false, 0, "")
			pdf.CellFormat(0, 10, "Timestamp: "+time.Now().Format("2006-01-02 15:04:05"), "", 1, "R", false, 0, "")
			pdf.Ln(8)
		}

		drawTableHeader := func() {
			pdf.SetFont("Arial", "B", 12)
			pdf.SetFillColor(headerBgColor.r, headerBgColor.g, headerBgColor.b)
			pdf.SetTextColor(headerTextColor.r, headerTextColor.g, headerTextColor.b)
			pdf.CellFormat(siteColWidth, 10, "Site", "1", 0, "C", true, 0, "")
			pdf.CellFormat(0, 10, "Details", "1", 1, "C", true, 0, "")
		}

		drawHeader()
		drawTableHeader()

		for siteName, siteURL := range vars.FoundSites[username] {
			estimatedHeight := 20.0 // Base height for site + URL
			if _, ok := vars.DeepScanResults[username][siteName]; ok {
				estimatedHeight += 25.0 // Add space for deep scan info
			}

			if pdf.GetY()+estimatedHeight > (pageHeight - pageMargin) {
				pdf.AddPage()
				drawHeader()
				drawTableHeader()
			}

			startY := pdf.GetY()
			pdf.SetX(pageMargin)

			pdf.SetFont("Arial", "B", 11)
			pdf.SetTextColor(primaryTextColor.r, primaryTextColor.g, primaryTextColor.b)
			pdf.MultiCell(siteColWidth, 10, siteName, "L", "L", false)

			// --- Details Column ---
			endYSite := pdf.GetY()
			pdf.SetY(startY)
			pdf.SetX(pageMargin + siteColWidth)

			pdf.SetFont("Courier", "", 9)
			pdf.SetTextColor(accentColor.r, accentColor.g, accentColor.b)
			pdf.MultiCell(0, 5, siteURL, "R", "L", false)
			pdf.SetX(pageMargin + siteColWidth)

			// Deep Scan Results
			if deepResult, ok := vars.DeepScanResults[username][siteName]; ok {
				pdf.Ln(2) // Add a little space
				val := reflect.ValueOf(deepResult)
				typ := val.Type()

				for i := 0; i < val.NumField(); i++ {
					field := val.Field(i)
					if field.IsNil() {
						continue
					}

					caser := cases.Title(language.English)

					fieldName := caser.String(strings.ReplaceAll(strings.Split(typ.Field(i).Tag.Get("json"), ",")[0], "_", " "))
					var fieldValue string

					switch f := field.Elem().Interface().(type) {
					case string:
						fieldValue = f
					case int:
						fieldValue = strconv.Itoa(f)
					case []string:
						fieldValue = strings.Join(f, ", ")
					case []vars.NonDefinedAction:
						var parts []string
						for _, action := range f {
							parts = append(parts, fmt.Sprintf("%s: %s", action.Name, action.Value))
						}
						fieldValue = strings.Join(parts, "; ")
					default:
						continue
					}

					if fieldValue != "" {
						pdf.SetX(pageMargin + siteColWidth + cardPadding) // Indent deep scan details

						// Field Name (e.g., "Real Name:")
						pdf.SetFont("Arial", "B", 9)
						pdf.SetTextColor(primaryTextColor.r, primaryTextColor.g, primaryTextColor.b)
						pdf.CellFormat(30, 5, fieldName+":", "", 0, "L", false, 0, "")

						// Field Value
						pdf.SetFont("Arial", "", 9)
						pdf.SetTextColor(secondaryTextColor.r, secondaryTextColor.g, secondaryTextColor.b)
						pdf.MultiCell(0, 5, fieldValue, "R", "L", false)
						pdf.SetX(pageMargin + siteColWidth)
					}
				}
			}

			// Determine the final height of the row block and draw the bottom border
			endYDetails := pdf.GetY()
			finalY := math.Max(endYSite, endYDetails)
			pdf.SetY(finalY)
			pdf.Line(pageMargin, finalY, pageWidth-pageMargin, finalY)
		}

		var buf bytes.Buffer
		err := pdf.Output(&buf)
		helpers.HandleErr(err)

		saveResultFile("pdf", username, buf.String())
	}
}

func saveResultFile(filetype string, username string, data string) {
	// helpers.V("Output Folder: %s", vars.OutputFolder)

	FileName := fmt.Sprintf("%s_results.%s", username, filetype)
	var FilePath string
	if len(vars.Usernames) > 1 {
		FilePath = filepath.Join(vars.OutputFolder, username)
		FilePath = filepath.Join(FilePath, FileName)
	} else {
		FilePath = filepath.Join(vars.OutputFolder, FileName)
	}
	f, err := os.Create(FilePath)
	helpers.HandleErr(err)

	defer func() { _ = f.Close() }()

	ByteData := []byte(data)

	_, err = f.Write(ByteData)
	helpers.HandleErr(err)
	// printer.Info("Save output file '%s'", FilePath)

}

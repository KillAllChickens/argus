package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/KillAllChickens/argus/internal/helpers"
	"github.com/KillAllChickens/argus/internal/io"
	"github.com/KillAllChickens/argus/internal/vars"

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
			"Username":  username,
			"Sites":     vars.FoundSites[username],
			"PFPs":      vars.FoundPFPs[username],
			"DeepScanEnabled": vars.DeepScanEnabled,
			"DeepScans": vars.DeepScanResults[username],
			"Timestamp": time.Now().Format("2006-01-02 15:04:05"),
			"Version":   vars.Version,
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
	for _, username := range vars.Usernames {
		pdf := gofpdf.New("P", "mm", "A4", "")
		defer pdf.Close()

		pdf.AddPage()
		pdf.SetFont("Arial", "B", 18)
		pdf.CellFormat(0, 15, "Argus Scan Results for "+username, "", 1, "C", false, 0, "")

		pdf.SetFont("Arial", "", 12)
		pdf.SetTextColor(40, 40, 40)
		pdf.CellFormat(0, 10, "Timestamp: "+time.Now().Format("2006-01-02 15:04:05"), "", 1, "C", false, 0, "")
		pdf.Ln(5)

		// Table header
		pdf.SetFont("Arial", "B", 14)
		pdf.SetFillColor(226, 232, 240) // light blue/gray fill
		pdf.CellFormat(45, 10, "Site", "1", 0, "C", true, 0, "")
		pdf.CellFormat(0, 10, "Profile URL & Deep Scan Details", "1", 1, "C", true, 0, "")

		pdf.SetFont("Arial", "", 12)
		pdf.SetTextColor(0, 0, 0)

		// Loop through sites and add their deep scan results.
		for siteName, siteURL := range vars.FoundSites[username] {
			pdf.CellFormat(45, 10, siteName, "1", 0, "L", false, 0, "")
			pdf.SetFont("Courier", "", 10)
			pdf.CellFormat(0, 10, siteURL, "1", 1, "L", false, 0, siteURL)
			pdf.SetFont("Arial", "", 12)

			// Check for and append deep scan results
			if deepResult, ok := vars.DeepScanResults[username][siteName]; ok {
				val := reflect.ValueOf(deepResult)
				typ := val.Type()

				for i := 0; i < val.NumField(); i++ {
					field := val.Field(i)
					if field.IsNil() {
						continue // Skip empty fields
					}

					fieldName := strings.ReplaceAll(strings.Split(typ.Field(i).Tag.Get("json"), ",")[0], "_", " ")
					var fieldValue string

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
						pdf.SetX(pdf.GetX() + 45) // Indent
						pdf.SetFont("Arial", "B", 10)
						pdf.CellFormat(35, 8, strings.Title(fieldName)+":", "L", 0, "L", false, 0, "")
						pdf.SetFont("Arial", "", 10)
						pdf.MultiCell(0, 8, fieldValue, "R", "L", false)
					}
				}
				// Draw bottom border for the entire row block
				pdf.Line(pdf.GetX(), pdf.GetY(), 200, pdf.GetY())

			}
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

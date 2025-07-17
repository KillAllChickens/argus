package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/KillAllChickens/argus/internal/helpers"
	"github.com/KillAllChickens/argus/internal/io"
	"github.com/KillAllChickens/argus/internal/vars"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

type outputJSONStruct struct {
	Username  string   `json:"username"`
	Timestamp string   `json:"timestamp"`
	Sites     []string `json:"sites"`
}

func OutputJSON() {
	for _, username := range vars.Usernames {
		data := outputJSONStruct{
			Username:  username,
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			Sites:     []string{},
		}

		for _, site := range vars.FoundSites[username] {
			data.Sites = append(data.Sites, site)
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
			"Username": username,
			"Sites":    vars.FoundSites[username],
			"PFPs":     vars.FoundPFPs[username],
		}

		funcMap := template.FuncMap{
			"index": func(m map[string]string, key string) string {
				return m[key]
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
		for n, site := range vars.FoundSites[username] {
			fullText += fmt.Sprintf("[+] %-14s => %-45s\n", n, site)
		}
		fullText += fmt.Sprintf("%d sites found for %s\n", len(vars.FoundSites[username]), username)
		saveResultFile("txt", username, fullText)
	}
}

func OutputPDF() {
	for _, username := range vars.Usernames {
		pdf := gofpdf.New("P", "mm", "A4", "")
		defer pdf.Close()

		// Add a page and set up fonts
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 18)
		pdf.CellFormat(0, 15, "Argus Scan Results for "+username, "", 1, "C", false, 0, "")

		pdf.SetFont("Arial", "B", 14)
		pdf.SetTextColor(40, 40, 40)
		pdf.Ln(5)

		// Table header
		pdf.SetFillColor(226, 232, 240) // light blue/gray fill
		pdf.CellFormat(60, 10, "Site", "1", 0, "C", true, 0, "")
		pdf.CellFormat(120, 10, "Profile URL", "1", 1, "C", true, 0, "")

		pdf.SetFont("Arial", "", 12)
		pdf.SetTextColor(0, 0, 0)

		// Loop through sites
		for n, site := range vars.FoundSites[username] {
			pdf.CellFormat(60, 10, n, "1", 0, "L", false, 0, "")
			// Assume the URL is just the site string here; adjust as needed.
			pdf.CellFormat(120, 10, site, "1", 1, "L", false, 0, site)
		}

		// Output PDF to buffer
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

	defer f.Close()

	ByteData := []byte(data)

	_, err = f.Write(ByteData)
	helpers.HandleErr(err)
	// printer.Info("Save output file '%s'", FilePath)

}

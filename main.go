package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	wappalayzer "github.com/projectdiscovery/wappalyzergo"
)

var (
	templateStr string
)

type RequestResponse struct {
	Request  string `json:"request"`
	Response string `json:"response"`
}

func init() {
	flag.StringVar(&templateStr, "output", "", `Template string for formatting AppInfo. Example: "{{join .Tech.Categories}} -> {{.Name}}\n{{.Tech.Description}}\n {{.Tech.Website}}\n{{.Tech.Icon}}\n{{.Tech.CPE}}\n-----------------------------\n"`)
}

func parseStdin() ([]byte, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, fmt.Errorf("error getting stdin stat: %v", err)
	}

	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil, errors.New("feed me with Caido HTTP RequestResponse using pipe")
	}

	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("can't read input: %v", err)
	}
	// Check for and remove BOM if present
	input = bytes.TrimPrefix(input, []byte("\ufeff"))

	var requestResponse RequestResponse
	if err := json.Unmarshal(input, &requestResponse); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	decodedResponse, err := base64.StdEncoding.DecodeString(requestResponse.Response)
	if err != nil {
		return nil, fmt.Errorf("error decoding base64: %v", err)
	}

	return decodedResponse, nil
}

func splitHTTPResponse(response []byte) (http.Header, []byte, error) {
	// Split response by headers and body
	parts := bytes.SplitN(response, []byte("\r\n\r\n"), 2)
	if len(parts) < 2 {
		return nil, nil, errors.New("invalid HTTP response format")
	}

	headers := make(http.Header)
	for i, line := range bytes.Split(parts[0], []byte("\r\n")) {
		if i == 0 {
			continue
		}
		sepIndex := bytes.IndexByte(line, ':')
		if sepIndex == -1 {
			return nil, nil, errors.New("invalid header format")
		}
		key := string(bytes.TrimSpace(line[:sepIndex]))
		value := string(bytes.TrimSpace(line[sepIndex+1:]))
		headers[key] = append(headers[key], value)
	}

	return headers, parts[1], nil
}

func analyze(headers map[string][]string, body []byte) (map[string]wappalayzer.AppInfo, error) {
	wappalayzerClient, err := wappalayzer.New()
	if err != nil {
		return nil, fmt.Errorf("error initializing Wappalyzer: %v", err)
	}

	appInfo := wappalayzerClient.FingerprintWithInfo(headers, body)
	return appInfo, nil
}

func joinList(items []string) string {
	return strings.Join(items, ",")
}

func formatOutput(techInfo map[string]wappalayzer.AppInfo, tmplStr string) (string, error) {
	var output strings.Builder

	for name, tech := range techInfo {
		if tmplStr == "" {
			output.WriteString(name + "\n")
			continue
		}

		var buffer bytes.Buffer
		parsedTemplateStr := bytes.ReplaceAll([]byte(tmplStr), []byte(`\n`), []byte("\n"))
		tmpl, err := template.New("techInfo").Funcs(template.FuncMap{
			"join": joinList,
		}).Parse(string(parsedTemplateStr))
		if err != nil {
			return "", fmt.Errorf("error parsing template: %v", err)
		}

		err = tmpl.Execute(&buffer, struct {
			Name string
			Tech wappalayzer.AppInfo
		}{name, tech})
		if err != nil {
			return "", fmt.Errorf("error executing template: %v", err)
		}

		output.WriteString(buffer.String())
	}

	return output.String(), nil
}

func main() {
	flag.Parse()

	decodedResponse, err := parseStdin()
	if err != nil {
		log.Fatalf("Error parsing stdin: %v", err)
	}

	headers, body, err := splitHTTPResponse(decodedResponse)
	if err != nil {
		log.Fatalf("Error splitting HTTP response: %v", err)
	}

	techInfo, err := analyze(headers, body)
	if err != nil {
		log.Fatalf("Error analyzing response: %v", err)
	}

	output, err := formatOutput(techInfo, templateStr)
	if err != nil {
		log.Fatalf("Error formatting output: %v", err)
	}

	fmt.Println(output)
}

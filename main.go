package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eiannone/keyboard"
	"github.com/go-resty/resty/v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Compliance struct {
	Name      string `json:"name"`
	Reference string `json:"reference"`
	Version   string `json:"version"`
}

type Finding struct {
	CheckedItems  int           `json:"checked_items"`
	Compliance    *[]Compliance `json:"compliance"`
	DashboardName string        `json:"dashboard_name"`
	Description   string        `json:"description"`
	DisplayPath   string        `json:"display_path"`
	FlaggedItems  int           `json:"flagged_items"`
	Items         []string      `json:"items"`
	Level         string        `json:"level"`
	Path          string        `json:"path"`
	Rationale     string        `json:"rationale"`
	References    []string      `json:"references"`
	Remediation   string        `json:"remediation"`
	Service       string        `json:"service"`
}

type Service struct {
	Findings map[string]Finding `json:"findings"`
}
type ScoutSuiteReport struct {
	Services map[string]Service `json:"services"`
}

type Prompt struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model         string   `json:"model"`
	Prompt        []Prompt `json:"prompt"`
	Temperature   float64  `json:"temperature"`
	MaxTokens     int      `json:"max_tokens"`
	StopSequences []string `json:"stop"`
}

type Response struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

func findReportFileName() (error, string) {
	baseDir := "scans"
	reportPattern := "scoutsuite_results_aws-*.js"
	var foundFiles []string

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing file or directory: %v\n", err)
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".js") {

			matchedFile, err := filepath.Match(reportPattern, info.Name())
			if err != nil {
				return err
			}
			if matchedFile {
				foundFiles = append(foundFiles, path)
			}
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	if len(foundFiles) == 0 {
		return errors.New("no files found"), ""
	}

	if len(foundFiles) > 1 {
		allFileNames := strings.Join(foundFiles, ", ")
		return errors.New(fmt.Sprintf("found too many files. Expected one: %s", allFileNames)), ""
	}

	return nil, foundFiles[0]
}

func loadFindings() []Finding {
	err, reportFile := findReportFileName()

	if err != nil {
		fmt.Printf("%v", err)
	}

	fmt.Printf("report name: %s\n", reportFile)

	reportBytes, err := ioutil.ReadFile(reportFile)
	if err != nil {
		panic(err)
	}

	// Remove `report =` before start of JSON
	reportBytes = reportBytes[bytes.Index(reportBytes, []byte("{")):]

	var report ScoutSuiteReport
	err = json.Unmarshal(reportBytes, &report)

	if err != nil {
		fmt.Println("Error unmarshalling json")
		panic(err)
	}

	var dangerFindings []Finding
	for _, service := range report.Services {
		for _, finding := range service.Findings {
			if finding.Level == "danger" {
				dangerFindings = append(dangerFindings, finding)
			}
		}
	}

	return dangerFindings
}

func printFinding(index int, total int, findings []Finding) {
	fmt.Printf("\033[2K")
	fmt.Printf("\033[%dA", 2)
	fmt.Printf("\033[2K")
	fmt.Printf("%s (%d of %d)\n", findings[index].Description, index+1, total)
	fmt.Printf("[j] for next, [k] for previous, [p] for info, [d] for devops, [q] to quit\n")
}

func printFindingAsString(f Finding) string {
	finding, err := json.MarshalIndent(f, "", "   ")
	if err != nil {
		fmt.Printf("%v", err)
		panic(err)
	}
	return string(finding)
}

func askQuestion(question string) {
	fmt.Println("Asking now!")
	fmt.Println("Asking now!")
	fmt.Println("Asking now!")
	fmt.Println("Asking now!")
	apiKey := os.Getenv("OPENAI_API_KEY")

	// Prepare the API request
	client := resty.New()
	request := Request{
		Model: "text-davinci-002",
		Prompt: []Prompt{
			{Role: "agent", Content: "Scout Suite report:"},
			{Role: "user", Content: question},
			{Role: "agent", Content: "Answer:"},
		},
		Temperature:   0.5,
		MaxTokens:     1024,
		StopSequences: []string{"\n"},
	}

	// Send the API request
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+apiKey).
		SetBody(request).
		Post("https://api.openai.com/v1/completions")
	if err != nil {
		fmt.Println("Error sending the request:", err)
		return
	}

	// Parse the API response
	var response Response
	err = json.Unmarshal(resp.Body(), &response)
	if err != nil {
		fmt.Println("Error parsing the response:", err)
		return
	}

	// Print the answer
	if len(response.Choices) > 0 {
		answer := response.Choices[0].Text
		// Remove leading and trailing whitespace
		answer = strings.TrimSpace(answer)
		fmt.Println("Answer:", answer)
	} else {
		fmt.Println("No answer found")
	}
	fmt.Println("End of question")
	fmt.Println("End of question")
	fmt.Println("End of question")
	fmt.Println("End of question")
	fmt.Println("End of question")
	fmt.Println("End of question")
}

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("OPENAI_API_KEY environment variable is required")
		return
	}
	findings := loadFindings()
	err := keyboard.Open()
	if err != nil {
		fmt.Printf("%v\n", err)
		panic(err)
	}
	defer keyboard.Close()

	currentIndex := 0
	total := len(findings)

	printFinding(currentIndex, total, findings)

	for {
		char, _, err := keyboard.GetKey()
		if err != nil {
			panic(err)
		}

		switch char {
		case 'j':
			currentIndex++
			if currentIndex >= total {
				currentIndex = total - 1
			}
		case 'k':
			currentIndex--
			if currentIndex < 0 {
				currentIndex = 0
			}
		case 'p':
			fmt.Printf("%s\n", printFindingAsString(findings[currentIndex]))
		case 'd':
			question := fmt.Sprintf("I got the following report from scoute suite... can you please write me a short step by step tutorial for fixing in my aws account: %s", printFindingAsString(findings[currentIndex]))
			askQuestion(question)
		case 'q': // Press 'q' to quit the program
			return
		default:
			continue
		}

		printFinding(currentIndex, total, findings)
	}
}

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"golang.org/x/net/html"
)

func getContent(pageURL string) (string, error) {
	client := &http.Client{
		Timeout:       10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return nil }, // Do not follow redirects
	}
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the full response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func getDirectoriesAndFiles(baseURL string, saveToFile bool) {
	visited := map[string]bool{}  // Keep track of visited URLs
	toVisit := []string{baseURL}  // Queue of URLs to visit
	var output []string

	for len(toVisit) > 0 {
		currentURL := toVisit[0]
		toVisit = toVisit[1:]

		if visited[currentURL] {
			continue
		}

		visited[currentURL] = true
		color.Cyan("Scanning: %s", currentURL)
		output = append(output, fmt.Sprintf("Scanning: %s", currentURL))

		content, err := getContent(currentURL)
		if err != nil {
			color.Red("Error accessing %s: %v", currentURL, err)
			output = append(output, fmt.Sprintf("Error accessing %s: %v", currentURL, err))
			continue
		}

		tokenizer := html.NewTokenizer(strings.NewReader(content))
		for {
			tt := tokenizer.Next()
			if tt == html.ErrorToken {
				break
			}

			token := tokenizer.Token()
			if token.Type == html.StartTagToken && token.Data == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						fullURL := resolveURL(currentURL, attr.Val)

						// Only visit URLs under the same domain
						if strings.HasPrefix(fullURL, baseURL) {
							if strings.HasSuffix(fullURL, "/") { // It's a directory
								if !visited[fullURL] && !contains(toVisit, fullURL) {
									color.Yellow("[DIR] %s", fullURL)
									output = append(output, fmt.Sprintf("[DIR] %s", fullURL))
									toVisit = append(toVisit, fullURL)
								}
							} else { // It's a file
								if !visited[fullURL] {
									color.Green("[FILE] %s", fullURL)
									output = append(output, fmt.Sprintf("[FILE] %s", fullURL))
								}
							}
						}
					}
				}
			}
		}
	}

	// Save output to file if the -s flag is provided
	if saveToFile {
		domain := strings.ReplaceAll(baseURL, "http://", "")
		domain = strings.ReplaceAll(domain, "https://", "")
		domain = strings.ReplaceAll(domain, "/", "")
		filename := fmt.Sprintf("%s.txt", domain)
		err := saveOutputToFile(filename, output)
		if err != nil {
			color.Red("Error saving to file: %v", err)
		} else {
			color.Green("Output saved to %s", filename)
		}
	}
}

func resolveURL(baseURL, href string) string {
	parsedBase, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	parsedHref, err := url.Parse(href)
	if err != nil {
		return ""
	}
	return parsedBase.ResolveReference(parsedHref).String()
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func saveOutputToFile(filename string, output []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range output {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}

func main() {
	// Define flags
	saveFlag := flag.Bool("s", false, "Save output to a file")
	flag.Parse()

	// Get domain from command line arguments
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: go run dir.go <domain>")
		fmt.Println("Use `go run dir.go -s <domain>` to save the output")
		os.Exit(1)
	}
	domain := args[0]

	// Add http if protocol is missing
	if !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
		domain = "http://" + domain
	}

	color.Cyan("\nStarting directory and file scan on %s\n", domain)
	getDirectoriesAndFiles(domain, *saveFlag)
}

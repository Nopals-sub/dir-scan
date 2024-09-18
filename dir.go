package main

import (
	"bufio"
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

func getDirectoriesAndFiles(baseURL string) {
	visited := map[string]bool{}  // Keep track of visited URLs
	toVisit := []string{baseURL}  // Queue of URLs to visit

	for len(toVisit) > 0 {
		currentURL := toVisit[0]
		toVisit = toVisit[1:]

		if visited[currentURL] {
			continue
		}

		visited[currentURL] = true
		color.Cyan("Scanning: %s", currentURL)

		content, err := getContent(currentURL)
		if err != nil {
			color.Red("Error accessing %s: %v", currentURL, err)
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
									toVisit = append(toVisit, fullURL)
								}
							} else { // It's a file
								if !visited[fullURL] {
									color.Green("[FILE] %s", fullURL)
								}
							}
						}
					}
				}
			}
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

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter domain (e.g., https://example.com): ")
	domain, _ := reader.ReadString('\n')
	domain = strings.TrimSpace(domain)

	color.Cyan("\nStarting directory and file scan on %s\n", domain)
	getDirectoriesAndFiles(domain)
}

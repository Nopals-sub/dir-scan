package main

import (
    "flag"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "os"
    "path/filepath"
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
    visited := make(map[string]bool)     // Keep track of visited URLs
    fileSeen := make(map[string]string)   // Track the directories where files were first encountered
    toVisit := []string{baseURL}          // Queue of URLs to visit

    var outputFile *os.File
    if saveToFile {
        outputPath := "output"
        os.MkdirAll(outputPath, os.ModePerm)
        domainName := strings.TrimPrefix(strings.TrimPrefix(baseURL, "https://"), "http://")
        domainName = strings.ReplaceAll(domainName, "/", "")
        filePath := filepath.Join(outputPath, domainName+".txt")
        var err error
        outputFile, err = os.Create(filePath)
        if err != nil {
            color.Red("Error creating output file: %v", err)
            return
        }
        defer outputFile.Close()
    }

    for len(toVisit) > 0 {
        currentURL := toVisit[0]
        toVisit = toVisit[1:]

        if visited[currentURL] {
            continue
        }

        visited[currentURL] = true
        color.Cyan("Scanning: %s", currentURL)
        writeToFile(outputFile, fmt.Sprintf("Scanning: %s\n", currentURL))

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
                                    writeToFile(outputFile, fmt.Sprintf("[DIR] %s\n", fullURL))
                                    toVisit = append(toVisit, fullURL)
                                }
                            } else { // It's a file
                                if firstDir, found := fileSeen[fullURL]; found {
                                    color.Magenta("[INFO] %s same like at %s", fullURL, firstDir)
                                    writeToFile(outputFile, fmt.Sprintf("[INFO] %s same like at %s\n", fullURL, firstDir))
                                } else {
                                    color.Green("[FILE] %s", fullURL)
                                    writeToFile(outputFile, fmt.Sprintf("[FILE] %s\n", fullURL))
                                    fileSeen[fullURL] = currentURL // Mark as seen with its directory
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

func writeToFile(file *os.File, content string) {
    if file != nil {
        file.WriteString(content)
    }
}

func main() {
    domain := flag.String("url", "", "The domain to scan (e.g., https://example.com)")
    saveToFile := flag.Bool("s", false, "Save output to a file")
    flag.Parse()

    if *domain == "" {
        color.Red("Please provide a domain using the -url flag.")
        return
    }

    color.Cyan("\nStarting directory and file scan on %s\n", *domain)
    getDirectoriesAndFiles(*domain, *saveToFile)
}

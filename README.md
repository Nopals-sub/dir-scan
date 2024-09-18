# Directory and File Scanner

This Go program scans directories and files on a specified domain. It recursively finds and lists directories and files, while filtering out duplicate results. It also handles both HTTP and HTTPS protocols and provides an option to resume or stop the scan if interrupted.

## Features

- **Recursive Scanning**: Scans directories and files found on a domain, and recursively scans through subdirectories.
- **Duplicate Filtering**: Ensures that duplicate files or directories are not displayed.
- **Protocol Handling**: Automatically adds `http://` if a domain is entered without `http://` or `https://`.

## Prerequisites

To run this project, you need:

- **Go** (Version 1.16 or higher)

## Installation

1. Clone the repository:

    ```bash
    git clone https://github.com/Nopals-sub/dir-scan.git
    ```

2. Navigate to the project directory:

    ```bash
    cd dir-scan
    ```

3. Build the project:

    ```bash
    go build dir.go
    ```

## Usage

Run the program and input a domain:

```bash
go run dir.go

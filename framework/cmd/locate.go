package internal

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

// LocateFiles searches for files and directories in the given path based on a search term
func LocateFiles(searchTerm string, searchPath string) ([]string, error) {
    var foundFiles []string

    // Walk through the directory and its subdirectories
    err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Check if the file or directory name contains the search term (case insensitive)
        if strings.Contains(strings.ToLower(info.Name()), strings.ToLower(searchTerm)) {
            foundFiles = append(foundFiles, path)
        }
        return nil
    })

    if err != nil {
        return nil, err
    }

    return foundFiles, nil
}


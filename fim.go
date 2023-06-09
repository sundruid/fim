package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var v, vv, h bool

var message = `

v.2

This is a multi-threaded application that performs a file integrity 
check using sha256 hashing and comparing since last scan. 

Hash changes will be recorded in FIMFILEA.OUT as TRUE after second 
scan if hashes do not match. FALSE for files that have not changed 
since last scan. 

You can exclude directories or files using the absolute file path 
in the exclude.config file. Directories listed will exclude all 
subdirectories. Will not follow symlinks.`

type fileInfo struct {
	path string
	hash string
}

func main() {
	flag.BoolVar(&v, "v", false, "Verbose mode (errors only)")
	flag.BoolVar(&vv, "vv", false, "Super verbose mode (all files processed)")
	flag.BoolVar(&h, "h", false, "Help")
	flag.Parse()

	if h {
		fmt.Println("Usage: fim [-v] [-vv] [-h]")
		fmt.Println("-v: Verbose mode (errors only)")
		fmt.Println("-vv: Super verbose mode (all files processed)")
		fmt.Println("-h: Help")
		fmt.Println(message)
		fmt.Println("")
		return
	}

	startTime := time.Now()

	var workerLimit = runtime.NumCPU()

	excludePaths := readExcludePaths("exclude.config")
	outFile := "FIMFILEA.OUT"
	tempFile := "FIMFILEA.TMP"
	previousScan := make(map[string]string)
	var previousScanMutex sync.Mutex

	fileExists := true
	if _, err := os.Stat(outFile); os.IsNotExist(err) {
		fileExists = false
	} else {
		previousScan = readPreviousScan(outFile)
	}

	file, err := os.Create(tempFile)
	if err != nil {
		if v || vv {
			fmt.Println("Error creating temporary file:", err)
		}
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	root := getRootDir()

	var wg sync.WaitGroup
	var writeMutex sync.Mutex

	// Create jobs channel
	jobs := make(chan string)

	// Launch worker goroutines
	for i := 0; i < workerLimit; i++ {
		go func() {
			for p := range jobs {
				processFile(p, &wg, &previousScanMutex, previousScan, &writeMutex, writer, fileExists)
			}
		}()
	}

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if v || vv {
				fmt.Printf("Error accessing path %q: %v\n", path, err)
			}
			return filepath.SkipDir
		}

		if shouldExclude(path, excludePaths) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		wg.Add(1)
		jobs <- path
		return nil
	})

	// Close the jobs channel and wait for all goroutines to finish
	close(jobs)
	wg.Wait()

	// Close the file and flush the writer explicitly for Windows
	writer.Flush()
	file.Close()

	err = os.Rename(tempFile, outFile)
	if err != nil {
		if os.IsExist(err) {
			fmt.Println("Destination file exists and might be in use. Trying to remove and replace...")

			removeErr := os.Remove(outFile)
			if removeErr != nil {
				if v || vv {
					fmt.Println("Error removing FIMFILEA.OUT:", removeErr)
				}
			} else {
				renameErr := os.Rename(tempFile, outFile)
				if renameErr != nil {
					if v || vv {
						fmt.Println("Error replacing FIMFILEA.OUT after removing:", renameErr)
					}
				}
			}
		} else {
			if v || vv {
				fmt.Println("Error replacing FIMFILEA.OUT:", err)
			}
		}
	}
	elapsedTime := time.Since(startTime)
	if v || vv {
		fmt.Printf("Total runtime: %s\n", elapsedTime)
	}
}

func processFile(path string, wg *sync.WaitGroup, previousScanMutex *sync.Mutex, previousScan map[string]string, writeMutex *sync.Mutex, writer *bufio.Writer, fileExists bool) {
	defer wg.Done()
	absPath := filepath.Clean(path)
	hash, err := calculateHash(absPath)
	if err != nil {
		if v || vv {
			fmt.Printf("Error calculating hash for file %s: %v\n", absPath, err)
		}
		return
	}

	if vv {
		fmt.Println(absPath)
	}

	changed := "FALSE"
	previousScanMutex.Lock()
	if prevHash, ok := previousScan[absPath]; ok {
		if prevHash != hash {
			changed = "TRUE"
		}
	} else {
		if !fileExists {
			changed = "FALSE"
		} else {
			changed = "TRUE"
		}
		previousScan[absPath] = hash
	}
	previousScanMutex.Unlock()

	writeMutex.Lock()
	_, err = writer.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\n", time.Now().Format(time.RFC3339), absPath, hash, changed))
	writeMutex.Unlock()
}

func getRootDir() string {
	switch osType := runtime.GOOS; osType {
	case "windows":
		return `C:\`
	case "darwin", "linux":
		return "/"
	default:
		if v || vv {
			fmt.Printf("Unsupported operating system: %s\n", osType)
		}
		return ""
	}
}

func readExcludePaths(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		if v || vv {
			fmt.Println("Error opening exclude.config:", err)
		}
		return nil
	}
	defer file.Close()

	paths := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			// Ignore empty lines or lines starting with '#'
			continue
		}
		paths = append(paths, line)
		if v || vv {
			fmt.Printf("Excluding path: %s\n", line)
		}
	}

	if err := scanner.Err(); err != nil {
		if v || vv {
			fmt.Println("Error reading exclude.config:", err)
		}
		return nil
	}

	return paths
}

func readPreviousScan(filename string) map[string]string {
	file, err := os.Open(filename)
	if err != nil {
		if v || vv {
			fmt.Println("Error opening FIMFILEA.OUT:", err)
		}
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	previousScan := make(map[string]string)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "\t")
		if len(parts) >= 3 {
			previousScan[parts[1]] = parts[2]
		}
	}
	return previousScan
}

func shouldExclude(path string, excludePaths []string) bool {
	for _, excludePath := range excludePaths {
		if strings.HasPrefix(path, excludePath) {
			return true
		}
	}
	return false
}

func calculateHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash), nil
}

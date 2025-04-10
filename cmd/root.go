package cmd

import (
	"fmt"
	"log"
	"os"
	"sync"

	W "app_cli/worker"
	"path/filepath"

	"github.com/spf13/cobra"
)

var rootcmd = &cobra.Command{
	Use:   "The_scanner",
	Short: "scans a folder or a file",
	Long:  "scans a folder or a file for specific use cases",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var scancmd = &cobra.Command{
	Use:     "folder_scanner",
	Aliases: []string{"scan"},
	Short:   "cli folder scanner to get all file details stored in a folder",
	Long:    "cli folder scanner to get all file details stored in a folder- name, size, datecreated",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			path := args[0]
			fmt.Println("Inside handler, path = ", path)

			entries, err := os.ReadDir(path)
			if err != nil {
				log.Fatal(err)
			}

			for _, e := range entries {
				fileInfo, err := os.Lstat(path + "\\" + e.Name())
				if err != nil {
					panic(err)
				}

				fmt.Println("Name : ", fileInfo.Name())
				fmt.Println("Size : ", fileInfo.Size(), "B")
				fmt.Println("Mode/permission : ", fileInfo.Mode())
				fmt.Println()

			}
		}
	},
}

var textcmd = &cobra.Command{
	Use:     "text_file_scanner",
	Aliases: []string{"filescan"},
	Short:   "cli to scan all text files to search for specific keywords",
	Long:    "cli to scan all text files to search for specific keywords defined in the config",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			path := args[0]
			var wg sync.WaitGroup
			fmt.Println("Inside handler, path = ", path)

			entries, err := os.ReadDir(path)
			if err != nil {
				log.Fatal(err)
			}

			var text_files []string
			var pdf_files []string
			for _, e := range entries {
				fpath := path + "\\" + e.Name()
				fileInfo, err := os.Lstat(fpath)
				if err != nil {
					panic(err)
				}

				name := fileInfo.Name()
				// fmt.Println("name", name)
				n := len(name)

				suff := name[n-4:]
				switch suff {
				case ".txt":
					text_files = append(text_files, name)
				case ".pdf":
					pdf_files = append(pdf_files, name)
				}
			}

			fmt.Println(text_files, pdf_files)

			maxGoroutines := 10
			guard := make(chan struct{}, maxGoroutines) // Semaphore to limit concurrency

			// Function to handle any file with appropriate worker
			processFile := func(fpath string, worker func(string)) {
				guard <- struct{}{} // block if guard channel is full
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer func() { <-guard }()
					worker(fpath)
				}()
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				// Process text files
				for _, fname := range text_files {
					fpath := filepath.Join(path, fname)
					processFile(fpath, W.Textworker)
				}
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				// Process PDF files
				for _, fname := range pdf_files {
					fpath := filepath.Join(path, fname)
					processFile(fpath, W.PDFWorker)
				}
			}()

			wg.Wait()

		}
	},
}

func Execute() {
	if err := rootcmd.Execute(); err != nil {
		fmt.Printf("Some error occured")
		os.Exit(1)
	}
}

func init() {
	rootcmd.AddCommand(scancmd)
	rootcmd.AddCommand(textcmd)
}

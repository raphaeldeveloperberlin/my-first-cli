/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// scrapeCmd represents the scrape command
var scrapeCmd = &cobra.Command{
	Use:   "scrape",
	Short: "Scraping the index.html file from a website.",
	Long:  `Scraping the index.html from a website involves retrieving the main HTML...`,
	Run: func(cmd *cobra.Command, args []string) {
		websiteUrl := ValidateFlagWebsiteUrl(cmd)
		location := ValidateLocationConfig()
		log.Println("Start to scrape website %v to location %v", websiteUrl, location)
		scrapeAndExtractHtml(websiteUrl, location)
	},
}

func init() {
	rootCmd.AddCommand(scrapeCmd)
	scrapeCmd.Flags().String("website-url", "", "Select a website-url")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scrapeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// scrapeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func ValidateLocationConfig() string {
	location := viper.GetViper().GetString("location")
	if location == "" {
		log.Println("Please add the location to .my-first-cli.yaml")
		os.Exit(1)
	}
	return location
}

func ValidateFlagWebsiteUrl(cmd *cobra.Command) string {
	websiteUrlFlag, err := cmd.Flags().GetString("website-url")
	if err != nil {
		log.Fatalf("Exception while reading flags - %v", err)
	}
	if websiteUrlFlag == "" {
		log.Println("Please add the flag website-url")
		os.Exit(1)
	}
	return websiteUrlFlag
}

func scrapeAndExtractHtml(websiteUrl, location string) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	var html string

	GetHtmlFromMainDom(websiteUrl, ctx, &html)
	saveToDisk(location, html)
}

func GetHtmlFromMainDom(webSiteUrl string, ctx context.Context, html *string) {
	if err := chromedp.Run(ctx,
		chromedp.Navigate(webSiteUrl),
		chromedp.WaitVisible(`html`, chromedp.ByQuery),
		chromedp.OuterHTML("html", html, chromedp.ByQuery),
	); err != nil {
		log.Fatal(err)
	}
}

func saveToDisk(filepath, filecontent string) {
	buffer := createBufferForHtmlString(filecontent)
	bufferToDestination(buffer, filepath)
}

func createBufferForHtmlString(filecontent string) *strings.Builder {
	buffer := new(strings.Builder)
	buffer.WriteString(filecontent + "\n")
	return buffer
}

func bufferToDestination(codeToCopyBuffer *strings.Builder, filepath string) {
	CreateFoldersIfNotExist(filepath)
	writeResultToDisk(codeToCopyBuffer, filepath)
}

func CreateFoldersIfNotExist(destinationFile string) {
	destinationFolder := filepath.Dir(destinationFile)
	err := os.MkdirAll(destinationFolder, os.ModePerm)
	if err != nil {
		log.Println("Error creating destination folder:", err)
		return
	}
}

func writeResultToDisk(toSaveContent *strings.Builder, pathFilename string) {

	openFile, err := os.Create(pathFilename)
	if err != nil {
		log.Print("File couldn't be created or truncated.")
	}
	defer func() {
		if err := openFile.Close(); err != nil {
			panic(err)
		}
	}()

	r := strings.NewReader(toSaveContent.String())
	w := bufio.NewWriter(openFile)

	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}

		if _, err := w.Write(buf[:n]); err != nil {
			panic(err)
		}
	}

	if err = w.Flush(); err != nil {
		panic(err)
	}
}

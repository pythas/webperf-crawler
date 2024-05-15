package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	strip "github.com/grokify/html-strip-tags-go"
)

type Section struct {
	Title       string
	Score       *float64
	ScoreMax    float64
	SubSections []SubSection
}

type SubSection struct {
	Title string
	Score *float64
}

const OutDir = "reports"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a site argument.")
		os.Exit(1)
	}
	site := os.Args[1]

	doc, err := htmlquery.LoadURL("https://webperf.se/site/" + site + "/")

	if err != nil {
		panic(err)
	}

	indices := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18}
	sections := []Section{}

	for _, index := range indices {
		section, err := htmlquery.QueryAll(doc, `//*[@id="content"]/div/div/div[2]/div/div/div/article[`+strconv.Itoa(index)+`]/div/div`)

		if err != nil {
			panic(err)
		}

		if section == nil {
			continue
		}

		str := htmlquery.OutputHTML(section[0], false)

		var lines []string

		for _, line := range strings.Split(str, "<br/>") {
			line = strip.StripTags(line)
			line = strings.Trim(line, " \n")

			if line == "" {
				continue
			}

			lines = append(lines, line)
		}

		currentSection := Section{}
		currentSection.Title = strings.Split(lines[0], "\n")[0]
		currentSection.SubSections = []SubSection{}

		regScore := regexp.MustCompile(`[+-]?((\d+\.?\d*)|(\.\d+)) av [+-]?((\d+\.?\d*)|(\.\d+))`)
		regSubScore := regexp.MustCompile(`\( [+-]?((\d+\.?\d*)|(\.\d+)) betyg \)`)

		for _, line := range lines {
			if strings.HasPrefix(line, "Betyg:") {
				if regScore.Match([]byte(line)) {
					match := regScore.FindStringSubmatch(line)

					score, err := strconv.ParseFloat(match[1], 64)

					if err != nil {
						panic(err)
					}

					scoreMax, err := strconv.ParseFloat(match[4], 64)

					if err != nil {
						panic(err)
					}

					currentSection.Score = &score
					currentSection.ScoreMax = scoreMax
				}
			}

			if strings.HasPrefix(line, "- ") {
				if regSubScore.Match([]byte(line)) {
					match := regSubScore.FindStringSubmatch(line)

					score, err := strconv.ParseFloat(match[1], 64)

					if err != nil {
						panic(err)
					}

					currentSection.SubSections = append(currentSection.SubSections, SubSection{
						Title: strings.Trim(line[2:len(line)-len(match[0])], " "),
						Score: &score,
					})
				} else {
					currentSection.SubSections = append(currentSection.SubSections, SubSection{
						Title: strings.Trim(line[2:], " "),
					})
				}
			}
		}

		sections = append(sections, currentSection)
	}

	if _, err := os.Stat(OutDir); os.IsNotExist(err) {
		os.Mkdir(OutDir, 0755)
	}

	var filename = OutDir + "/" + site + "_" + time.Now().Format("2006_01_02_15_04") + ".json"

	createAndWriteFile(filename, sections)

	println("Report created for " + site + ": " + filename)
}

func createAndWriteFile(filename string, sections []Section) {
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	if err := encoder.Encode(sections); err != nil {
		panic(err)
	}
}

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	_ "embed" // Import the embed package
)

var (
	ignoreMode   bool
	ignoreOwner  bool
	ignoreSize   bool
	ignoreTime   bool
	ignoreDate   bool
	ignoreRemain bool
	templateFile string
	debugMode    bool
)

//go:embed output1.tmpl
var templateContent string

func main() {
	flag.BoolVar(&ignoreMode, "no-mode", false, "Ignore the mode field")
	flag.BoolVar(&ignoreOwner, "no-owner", false, "Ignore the owner field")
	flag.BoolVar(&ignoreSize, "no-size", false, "Ignore the size field")
	flag.BoolVar(&ignoreTime, "no-time", false, "Ignore the time field")
	flag.BoolVar(&ignoreDate, "no-date", false, "Ignore the date field")
	flag.BoolVar(&ignoreRemain, "no-remain", false, "Ignore the remaining field")
	flag.BoolVar(&debugMode, "debug", false, "Enable debug mode")
	flag.StringVar(&templateFile, "template", "", "Path to a template file for output")

	flag.Parse()

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for EOF
		if line == "" {
			break
		}

		fields := strings.Fields(line)
		if len(fields) < 6 {
			fmt.Println("Invalid input line:", line)
			continue
		}

		mode := fields[0]
		owner := fields[1]

		// Convert the size field from string to integer
		sizeStr := fields[2]
		sizeInt, err := strconv.Atoi(sizeStr)
		if err != nil {
			fmt.Println("Error converting size to int:", err)
			continue
		}

		date := fields[3]
		timeValue := fields[4]
		remaining := strings.Join(fields[5:], " ")

		dateTimeStr := fmt.Sprintf("%s %s", date, timeValue)
		timeStamp, err := time.Parse("2006-01-02 15:04", dateTimeStr)
		if err != nil {
			fmt.Println("Error parsing timeStamp:", err)
			continue
		}

		if debugMode {
			log.Printf("Parsing mode field: %s", mode)
			log.Printf("Parsing owner field: %s", owner)
			log.Printf("Parsing size field: %s", sizeStr)
			log.Printf("Parsing date field: %s", date)
			log.Printf("Parsing time field: %s", timeValue)
			log.Printf("Parsing remaining field: %s", remaining)
		}

		tmplString := templateContent

		// Use the specified template file if provided
		if templateFile != "" {
			tmplBytes, err := os.ReadFile(templateFile)
			if err != nil {
				fmt.Println("Error reading template file:", err)
				continue
			}
			tmplString = string(tmplBytes)
		}

		// Define a custom template function to repeat a string
		funcMap := template.FuncMap{
			"repeat": strings.Repeat,
		}

		// Parse the template
		tmpl, err := template.New("output").Funcs(funcMap).Parse(tmplString)
		if err != nil {
			fmt.Println("Error parsing template:", err)
			continue
		}

		// Data for the template
		data := struct {
			IgnoreMode   bool
			IgnoreOwner  bool
			IgnoreSize   bool
			IgnoreTime   bool
			IgnoreRemain bool
			Mode         string
			Owner        string
			Size         int // Change Size to int
			TimeStamp    time.Time
			Remaining    string
		}{
			IgnoreMode:   ignoreMode,
			IgnoreOwner:  ignoreOwner,
			IgnoreSize:   ignoreSize,
			IgnoreTime:   ignoreTime,
			IgnoreRemain: ignoreRemain,
			Mode:         mode,
			Owner:        owner,
			Size:         sizeInt, // Use the converted integer value
			TimeStamp:    timeStamp,
			Remaining:    remaining,
		}

		// Execute the template
		err = tmpl.Execute(os.Stdout, data)
		if err != nil {
			fmt.Println("Error executing template:", err)
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input:", err)
	}
}

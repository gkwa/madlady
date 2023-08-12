package main

import (
	"archive/tar"
	"flag"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const templateStr = `
{{- .ModeStr }} {{ .UIDName }}/{{ .GIDName }}  {{ .Size | printf "%8d" }}{{ if .ShowTimestamp }} {{ .MTime }}{{ end }} {{ .Name }}
`

type TarEntry struct {
	ModeStr       string
	UIDName       string
	GIDName       string
	Size          int64
	MTime         string
	Name          string
	ShowTimestamp bool // New field to control timestamp visibility
}

func convertBytesToStr(data interface{}) interface{} {
	switch v := data.(type) {
	case []byte:
		return string(v)
	case map[interface{}]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			result[convertBytesToStr(key).(string)] = convertBytesToStr(value)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = convertBytesToStr(item)
		}
		return result
	default:
		return data
	}
}

func modeToStr(mode os.FileMode) string {
	permStr := ""
	permStr += "d" // Directory indicator
	permStr += "r" // Owner readable
	permStr += "w" // Owner writable
	permStr += "x" // Owner executable
	permStr += "r" // Group readable
	permStr += "w" // Group writable
	permStr += "x" // Group executable
	permStr += "r" // Others readable
	permStr += "w" // Others writable
	permStr += "x" // Others executable

	// Check each permission bit and set the appropriate character
	if mode&os.ModeDir == 0 {
		permStr = strings.Replace(permStr, "d", "-", 1)
	}
	if mode&os.FileMode(0o400) == 0 {
		permStr = strings.Replace(permStr, "r", "-", 2)
	}
	if mode&os.FileMode(0o200) == 0 {
		permStr = strings.Replace(permStr, "w", "-", 2)
	}
	if mode&os.FileMode(0o100) == 0 {
		permStr = strings.Replace(permStr, "x", "-", 2)
	}
	if mode&os.FileMode(0o040) == 0 {
		permStr = strings.Replace(permStr, "r", "-", 5)
	}
	if mode&os.FileMode(0o020) == 0 {
		permStr = strings.Replace(permStr, "w", "-", 5)
	}
	if mode&os.FileMode(0o010) == 0 {
		permStr = strings.Replace(permStr, "x", "-", 5)
	}
	if mode&os.FileMode(0o004) == 0 {
		permStr = strings.Replace(permStr, "r", "-", 8)
	}
	if mode&os.FileMode(0o002) == 0 {
		permStr = strings.Replace(permStr, "w", "-", 8)
	}
	if mode&os.FileMode(0o001) == 0 {
		permStr = strings.Replace(permStr, "x", "-", 8)
	}

	return permStr
}

func formatDatetime(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format("2006-01-02 15:04")
}

func getUsername(uid int) string {
	user, err := user.LookupId(strconv.Itoa(uid))
	if err == nil {
		return user.Username
	}
	return strconv.Itoa(uid)
}

func getGroupname(gid int) string {
	group, err := user.LookupGroupId(strconv.Itoa(gid))
	if err == nil {
		return group.Name
	}
	return strconv.Itoa(gid)
}

func parseTarToTemplate(tarFilename string, showTimestamp bool) (string, error) {
	tarFile, err := os.Open(tarFilename)
	if err != nil {
		return "", err
	}
	defer tarFile.Close()

	tarReader := tar.NewReader(tarFile)
	tarEntries := []TarEntry{}

	for {
		header, err := tarReader.Next()
		if err != nil {
			break
		}

		fileInfo := TarEntry{
			ModeStr:       modeToStr(header.FileInfo().Mode()),
			UIDName:       getUsername(int(header.Uid)),
			GIDName:       getGroupname(int(header.Gid)),
			Size:          header.Size,
			Name:          convertBytesToStr(header.Name).(string),
			ShowTimestamp: showTimestamp,
		}

		if showTimestamp {
			fileInfo.MTime = formatDatetime(header.ModTime.Unix())
		}

		tarEntries = append(tarEntries, fileInfo)
	}

	tmpl, err := template.New("entry").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var output strings.Builder
	for _, entry := range tarEntries {
		err = tmpl.Execute(&output, entry)
		if err != nil {
			return "", err
		}
	}

	return output.String(), nil
}

func main() {
	tarPath := flag.String("path", "", "Path to the tar file")
	hideTimestamp := flag.Bool("timestamp", false, "Exclude timestamp from the output")

	flag.Parse()

	if *tarPath == "" {
		fmt.Println("Please provide a valid path to the tar file.")
		return
	}

	templateOutput, err := parseTarToTemplate(*tarPath, *hideTimestamp)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(templateOutput)
}

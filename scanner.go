package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"os/exec"
	"strings"
)

type TagReaderCh struct {
	Tags <-chan *Tag
	errs <-chan error
	done <-chan struct{}
}

func scannerRun() *TagReaderCh {
	tagChan := make(chan *Tag)
	errsChan := make(chan error)

	var currentTable string
	var tagData *Tag
	var tagReader *TagReader

	args := "-listx"
	cmd := exec.Command("exiftool", strings.Split(args, " ")...)

	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	go func() {
		for scanner.Scan() {
			l := scanner.Text()

			if strings.Contains(l, "<taginfo>") {
				continue
			}

			// Start reading table
			if strings.Contains(l, "<table") {
				currentTable, _ = readTableData(l)
				continue
			}

			// End a table
			if strings.Contains(l, "</table") {
				currentTable = ""
				tagReader = nil
				tagData = nil
				continue
			}

			// Start reading a tag
			if strings.Contains(l, "<tag") {
				tagReader = NewTagReader(currentTable)
				tagReader.Begin(l)
				continue
			}

			// Parse a completed tag
			if strings.Contains(l, "</tag>") {
				tagReader.AddLine(l)
				tagData, _ = tagReader.Parse()
				tagChan <- tagData
				tagReader = nil
				tagData = nil
			}

			if tagReader != nil {
				tagReader.AddLine(l)
			}

			fmt.Println(currentTable)
			fmt.Println(tagData)
		}
		cmd.Wait()
		close(tagChan)
	}()
	return &TagReaderCh{
		Tags: tagChan,
		errs: errsChan,
	}
}

func readTableData(inp string) (string, error) {
	type Table struct {
		Name string `xml:"name,attr"`
	}
	xmlTag := []byte(fmt.Sprintf("%s </table>", inp))
	table := &Table{}
	if err := xml.Unmarshal(xmlTag, table); err != nil {
		return "", err
	}
	return table.Name, nil
}

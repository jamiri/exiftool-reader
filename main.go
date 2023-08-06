package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

type TagInfo struct {
	XMLName xml.Name `xml:"taginfo"`
	Table   []Table  `xml:"table"`
}

type Table struct {
	Name string `xml:"name,attr"`

	Desc Desc  `xml:"desc"`
	Tags []Tag `xml:"tag"`
}

type Desc struct {
	Lang  string `xml:"lang,attr"`
	Value string `xml:",chardata"`
}

type Tag struct {
	Name      string `xml:"name,attr"`
	Type      string `xml:"type,attr"`
	Writable  string `xml:"writable,attr"`
	Descs     []Desc `xml:"desc"`
	TableName string
}

func main() {
	http.HandleFunc("/tags", getTags)
	http.ListenAndServe("localhost:8889", nil)
}

func getTags(w http.ResponseWriter, r *http.Request) {
	jsonData := "{\"tags\": [\n"
	w.Header().Set("Content-Type", "application/octet-stream")
	io.Copy(w, strings.NewReader(jsonData))
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	controlCh := scannerRun()

	//select {
	//case table := <-controlCh.Tables:
	//	for tag := range controlCh.Tags {
	//		wData := fmt.Sprintf("{\n\"writeable\": %s,\n", tag.Writable)
	//		wData = wData + fmt.Sprintf("\"path\": %s:%s,\n", table, tag.Name)
	//		wData = wData + fmt.Sprintf("\"group\": %s,\n", table)
	//		wData = wData + fmt.Sprintf("\"description\":{\n")
	//		tagAppendix := ","
	//		for i, desc := range tag.Descs {
	//			if i == len(tag.Descs)-1 {
	//				tagAppendix = ""
	//			}
	//			wData = wData + fmt.Sprintf("\"%s\": \"%s\"%s\n", desc.Lang, desc.Value, tagAppendix)
	//		}
	//		wData = wData + "},\n"
	//		tagAppendix = ","
	//		io.Copy(w, strings.NewReader(wData))
	//		if f, ok := w.(http.Flusher); ok {
	//			f.Flush()
	//		}
	//	}
	//}

	//for table := range controlCh.Tables {
	for tag := range controlCh.Tags {
		wData := fmt.Sprintf("{\n\"writeable\": %s,\n", tag.Writable)
		wData = wData + fmt.Sprintf("\"path\": %s:%s,\n", tag.TableName, tag.Name)
		wData = wData + fmt.Sprintf("\"group\": %s,\n", tag.TableName)
		wData = wData + fmt.Sprintf("\"description\":{\n")
		tagAppendix := ","
		for i, desc := range tag.Descs {
			if i == len(tag.Descs)-1 {
				tagAppendix = ""
			}
			wData = wData + fmt.Sprintf("\"%s\": \"%s\"%s\n", desc.Lang, desc.Value, tagAppendix)
		}
		wData = wData + "},\n"
		tagAppendix = ","
		io.Copy(w, strings.NewReader(wData))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
	<-controlCh.done
}

type TagReaderCh struct {
	//Tables <-chan string
	Tags <-chan *Tag
	errs <-chan error
	done <-chan struct{}
}

func scannerRun() *TagReaderCh {
	//tableChan := make(chan string)
	tagChan := make(chan *Tag)
	errsChan := make(chan error)
	done := make(chan struct{})

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
				//tableChan <- currentTable
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
	}()
	return &TagReaderCh{
		//Tables: tableChan,
		Tags: tagChan,
		errs: errsChan,
		done: done,
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

func readTagsData(inp string) (string, error) {
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

type TagReader struct {
	Data      string
	TableName string
}

func NewTagReader(tableName string) *TagReader {
	return &TagReader{
		TableName: tableName,
	}
}

func (tReader *TagReader) Begin(line string) {
	tReader.Data = line
}

func (tReader *TagReader) AddLine(line string) {
	tReader.Data = fmt.Sprintf("%s\n%s", tReader.Data, line)
}

func (tReader *TagReader) Parse() (*Tag, error) {
	xmlData := []byte(tReader.Data)
	tagData := &Tag{}
	if err := xml.Unmarshal(xmlData, tagData); err != nil {
		return nil, err
	}
	tagData.TableName = tReader.TableName
	return tagData, nil
}

//func (tReader *TagReader) End() {
//
//}

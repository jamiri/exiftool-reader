package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"os/exec"
	"strings"
)

const sample = `<?xml version='1.0' encoding='UTF-8'?>
<!-- Generated by Image::ExifTool 11.30 -->
<taginfo>
<table name='AFCP::Main' g0='AFCP' g1='AFCP' g2='Other'>
 <desc lang='en'>AFCP</desc>
 <tag id='Nail' name='ThumbnailImage' type='?' writable='false' g2='Preview'>
  <desc lang='en'>Thumbnail Image</desc>
  <desc lang='cs'>Náhled</desc>
  <desc lang='de'>Miniaturbild</desc>
  <desc lang='es'>Miniatura</desc>
</tag>
</table>
</taginfo>
`

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
	Name     string `xml:"name,attr"`
	Type     string `xml:"type,attr"`
	Writable string `xml:"writable,attr"`
	Descs    []Desc `xml:"desc"`
}

func main() {
	scannerRun()

	//tInfo := &TagInfo{
	//	//Table: make([]Table, 0),
	//}
	//_ = xml.Unmarshal([]byte(sample), tInfo)
	//fmt.Println(tInfo)
}

func scannerRun() {
	args := "-listx"
	cmd := exec.Command("exiftool", strings.Split(args, " ")...)

	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)

	var currentTable string
	var tagData *Tag
	var tagReader *TagReader
	for scanner.Scan() {
		l := scanner.Text()

		if strings.Contains(l, "<table") {
			currentTable, _ = readTableData(l)
			continue
		}

		if strings.Contains(l, "<tag") {
			tagReader = &TagReader{}
			tagReader.Begin(l)
			continue
		}

		if strings.Contains(l, "</tag>") {
			tagReader.AddLine(l)
			tagData, _ = tagReader.Parse()
			tagReader = nil
		}

		if tagReader != nil {
			tagReader.AddLine(l)
		}

		fmt.Println(currentTable)
		fmt.Println(tagData)
	}
	cmd.Wait()
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
	Data string
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
	return tagData, nil
}

//func (tReader *TagReader) End() {
//
//}

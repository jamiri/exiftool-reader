package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"os/exec"
	"strings"
)

const sample = `
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
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`

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
	tInfo := &TagInfo{
		//Table: make([]Table, 0),
	}
	_ = xml.Unmarshal([]byte(sample), tInfo)
	fmt.Println(tInfo)
}

func scannerRun() {
	args := "-listx"
	cmd := exec.Command("exiftool", strings.Split(args, " ")...)

	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}
	cmd.Wait()
}

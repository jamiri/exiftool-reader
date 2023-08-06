package main

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

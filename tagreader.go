package main

import (
	"encoding/xml"
	"fmt"
)

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

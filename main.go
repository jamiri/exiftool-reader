package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func main() {
	http.HandleFunc("/tags", getTags)
	fmt.Println("Server running on port 8889")
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

	for tag := range controlCh.Tags {
		wData := fmt.Sprintf(",\n{\n\"writeable\": %s,\n", tag.Writable)
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
		wData = wData + "}"
		tagAppendix = ","
		io.Copy(w, strings.NewReader(wData))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
	io.Copy(w, strings.NewReader("\n]\n}"))
}

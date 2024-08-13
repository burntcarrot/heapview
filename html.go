package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"regexp"
	"strings"

	"github.com/burntcarrot/heaputil"
	"github.com/burntcarrot/heaputil/record"
)

type templateData struct {
	RecordTypes []RecordInfo
	Records     []heaputil.RecordData
	GraphData   string
}

func GenerateHTML(records []heaputil.RecordData, graphContent string) (string, error) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		return "", err
	}

	data := templateData{
		RecordTypes: GetUniqueRecordTypes(records),
		Records:     records,
		GraphData:   graphContent,
	}

	var htmlBuilder strings.Builder
	err = tmpl.Execute(&htmlBuilder, data)
	if err != nil {
		return "", err
	}

	return htmlBuilder.String(), nil
}

func GenerateGraph(rd *bufio.Reader) (string, error) {
	err := record.ReadHeader(rd)
	if err != nil {
		return "", err
	}

	nodes := []map[string]interface{}{}
	links := []map[string]interface{}{}
	nodeMap := make(map[uint64]int)

	var dumpParams *record.DumpParamsRecord
	counter := 0

	for {
		r, err := record.ReadRecord(rd)
		if err != nil {
			break
		}

		_, isEOF := r.(*record.EOFRecord)
		if isEOF {
			break
		}

		dp, isDumpParams := r.(*record.DumpParamsRecord)
		if isDumpParams {
			dumpParams = dp
		}

		obj, isObj := r.(*record.ObjectRecord)
		if !isObj {
			continue
		}

		name, address := ParseNameAndAddress(r.Repr())
		nodeLabel := fmt.Sprintf("[%s] %s", name, address)

		if _, exists := nodeMap[obj.Address]; !exists {
			nodeMap[obj.Address] = counter
			nodes = append(nodes, map[string]interface{}{
				"id":      counter,
				"label":   nodeLabel,
				"address": obj.Address,
			})
			counter++
		}

		p, isParent := r.(record.ParentGuard)
		if isParent {
			_, outgoing := record.ParsePointers(p, dumpParams)
			for i := 0; i < len(outgoing); i++ {
				if outgoing[i] != 0 {
					if targetIndex, exists := nodeMap[outgoing[i]]; exists {
						links = append(links, map[string]interface{}{
							"source": nodeMap[obj.Address],
							"target": targetIndex,
						})
					}
				}
			}
		}
	}

	graphData := map[string]interface{}{
		"nodes": nodes,
		"links": links,
	}

	jsonData, err := json.Marshal(graphData)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func ParseNameAndAddress(input string) (name, address string) {
	// Define a regular expression pattern to match the desired format
	// The pattern captures the node name (before " at address") and the address.
	re := regexp.MustCompile(`^(.*?) at address (0x[0-9a-fA-F]+).*?$`)

	// Find the submatches in the input string
	matches := re.FindStringSubmatch(input)

	// If there are no matches, return empty strings for both name and address
	if len(matches) != 3 {
		return "", ""
	}

	// The first submatch (matches[1]) contains the node name, and the second submatch (matches[2]) contains the address.
	return matches[1], matches[2]
}

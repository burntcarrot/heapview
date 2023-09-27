package main

import (
	"bufio"
	"fmt"
	"html/template"
	"regexp"
	"strings"

	"github.com/burntcarrot/heaputil"
	"github.com/burntcarrot/heaputil/record"
)

type templateData struct {
	RecordTypes     []RecordInfo
	Records         []heaputil.RecordData
	GraphVizContent string
}

func GenerateHTML(records []heaputil.RecordData, graphContent string) (string, error) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		return "", err
	}

	data := templateData{
		RecordTypes:     GetUniqueRecordTypes(records),
		Records:         records,
		GraphVizContent: graphContent,
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

	var dotContent strings.Builder

	// Write DOT file header
	dotContent.WriteString("digraph GoHeapDump {\n")

	// Create the "heap" node as a cluster
	dotContent.WriteString("  subgraph cluster_heap {\n")
	dotContent.WriteString("    label=\"Heap\";\n")
	dotContent.WriteString("    style=dotted;\n")

	var dumpParams *record.DumpParamsRecord
	counter := 0

	for {
		r, err := record.ReadRecord(rd)
		if err != nil {
			return dotContent.String(), err
		}

		_, isEOF := r.(*record.EOFRecord)
		if isEOF {
			break
		}

		dp, isDumpParams := r.(*record.DumpParamsRecord)
		if isDumpParams {
			dumpParams = dp
		}

		// Filter out objects. If the record isn't of the type Object, ignore.
		_, isObj := r.(*record.ObjectRecord)
		if !isObj {
			continue
		}

		// Create a DOT node for each record
		nodeName := fmt.Sprintf("Node%d", counter)
		counter++
		name, address := ParseNameAndAddress(r.Repr())
		nodeLabel := fmt.Sprintf("[%s] %s", name, address)

		// Write DOT node entry within the "heap" cluster
		s := fmt.Sprintf("    %s [label=\"%s\"];\n", nodeName, nodeLabel)
		dotContent.WriteString(s)

		// Check if the record has pointers
		p, isParent := r.(record.ParentGuard)
		if isParent {
			_, outgoing := record.ParsePointers(p, dumpParams)
			for i := 0; i < len(outgoing); i++ {
				if outgoing[i] != 0 {
					childNodeName := fmt.Sprintf("Pointer0x%x", outgoing[i])

					// Create an edge from the current record to the child record
					s := fmt.Sprintf("    %s -> %s;\n", nodeName, childNodeName)
					dotContent.WriteString(s)
				}
			}
		}
	}

	// Close the "heap" cluster
	dotContent.WriteString("  }\n")

	// Write DOT file footer
	dotContent.WriteString("}\n")

	return dotContent.String(), nil
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

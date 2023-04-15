package cmd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Demonstration of the working with the Erigon database remotely on the example of getting information
// about past reorganisation of the chain

func (uih *UiHandler) findReorgs(w http.ResponseWriter, requestChannel chan *NodeRequest) {
	//w.Header().Set("Content-Type", "application/octet-stream")
	start := time.Now()
	// First, fetch list of DB paths
	success, result := uih.fetch("/db/list\n", requestChannel)
	if !success {
		fmt.Fprintf(w, "Fetching list of db paths: %s", result)
		return
	}
	lines := strings.Split(result, "\n")
	if len(lines) == 0 || !strings.HasPrefix(lines[0], successLine) {
		fmt.Fprintf(w, "Incorrect response (first line needs to be SUCCESS): %v", lines)
		return
	}
	var chaindataPath string
	for _, line := range lines[1:] {
		if strings.HasSuffix(line, "/chaindata") {
			chaindataPath = line
		}
	}
	if chaindataPath == "" {
		fmt.Fprintf(w, "DB path chaindata not found: %v", lines)
		return
	}
	// Go through "Header" table and look for entries with the same block number but different hashes
	var prevK []byte
	reorgCount := 0
	rc, err := NewRemoteCursor(chaindataPath, "Header", requestChannel, nil)
	if err != nil {
		fmt.Fprintf(w, "Create remote cursor: %v", err)
		return
	}
	var k []byte
	var e error
	var count int
	for k, _, e = rc.Next(); e == nil && k != nil; k, _, e = rc.Next() {
		if len(k) >= 8 && len(prevK) >= 8 && bytes.Equal(k[:8], prevK[:8]) {
			bn := binary.BigEndian.Uint64(k[:8])
			fmt.Fprintf(w, "<div class=\"block\">%d</div>\n", bn)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			reorgCount++
		}
		prevK = k
		count++
		if count%1000 == 0 {
			fmt.Fprintf(w, "<div class=\"progress\"></div>\n")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
	if e != nil {
		fmt.Fprintf(w, "Process remote cursor line: %v\n", e)
		return
	}
	fmt.Fprintf(w, "Reorg count: %d, produced in %s\n", reorgCount, time.Since(start))
}

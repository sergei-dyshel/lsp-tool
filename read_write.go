package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type filterMode int

const (
	enableFilter filterMode = iota
	disableFilter
	noFilter
)

func stdoutWrite(b []byte) {
	osStdout := bufio.NewWriter(os.Stdout)
	_, e := fmt.Fprintf(osStdout, "Content-Length: %d\r\n\r\n", len(b))
	panicIfError(e)
	_, e = osStdout.Write(b)
	panicIfError(e)
	osStdout.Flush()
}

func stdoutReader(lsStdout io.ReadCloser, mode filterMode, providers []string) {
	// Build scanner which will process LSP messages.
	scanner := bufio.NewScanner(lsStdout)
	scanner.Split(jsonRpcSplitFunc)
	osStdout := bufio.NewWriter(os.Stdout)
	for scanner.Scan() {
		log.Printf("response: %s", scanner.Text())
		var f interface{}
		err := json.Unmarshal(scanner.Bytes(), &f)
		panicIfError(err)
		jsonRoot, had := f.(map[string]interface{})
		if !had {
			stdoutWrite(scanner.Bytes())
			continue
		}
		jsonResult, had := jsonRoot["result"].(map[string]interface{})
		if !had {
			stdoutWrite(scanner.Bytes())
			continue
		}
		jsonCapabilities, had := jsonResult["capabilities"].(map[string]interface{})
		if !had {
			stdoutWrite(scanner.Bytes())
			continue
		}

		enabled := []string{}
		disabled := []string{}

		for k := range jsonCapabilities {
			if !strings.HasSuffix(k, "Provider") {
				continue
			}
			rawName := strings.TrimSuffix(k, "Provider")
			contains := indexOf(rawName, providers) >= 0

			if (mode == disableFilter && contains) ||
				(mode == enableFilter && !contains) {
				jsonCapabilities[k] = false
				disabled = append(disabled, rawName)
			} else {
				enabled = append(enabled, rawName)
			}
		}
		log.Printf("Enabled capabilities: %v", enabled)
		log.Printf("Disabled capabilities: %v", disabled)

		// Serialize and write message
		b, e := json.Marshal(f)
		panicIfError(e)
		stdoutWrite(b)

		break
	}
	panicIfError(scanner.Err())

	// Write the rest of the content blindly
	var buffer [1024]byte
	for {
		n, e := lsStdout.Read(buffer[:])
		if e != nil {
			break
		}

		logDebug(string(buffer[:n]))

		_, e = osStdout.Write(buffer[:n])
		osStdout.Flush()
		if e != nil {
			break
		}
	}

	// Close stdout.
	panicIfError(lsStdout.Close())
}

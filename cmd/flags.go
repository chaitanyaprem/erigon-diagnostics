package cmd

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type Flags struct {
	Success     bool
	Error       string
	FlagPayload map[string]string
}

func processFlags(w http.ResponseWriter, templ *template.Template, success bool, result string, versions Versions) {

	if !versions.Success {
		fmt.Fprintf(w, "Unable to process flag due to inability to get node version: %s", versions.Error)
		return
	}

	if versions.NodeVersion < 2 {
		fmt.Fprintf(w, "Flags only support version >= 2. Node version: %d", versions.NodeVersion)
		return
	}

	var flags Flags
	flags.FlagPayload = make(map[string]string)
	if success {
		lines := strings.Split(result, "\n")
		if len(lines) > 0 && strings.HasPrefix(lines[0], successLine) {
			flags.Success = true

			for _, line := range lines[1:] {
				if len(line) > 0 {
					flagName, flagValue, found := strings.Cut(line, "=")
					if !found {
						flags.Error = fmt.Sprintf("fail to parse line %s", line)
						flags.Success = false
					} else {
						flags.FlagPayload[flagName] = flagValue
					}
				}
			}
		} else {
			flags.Error = fmt.Sprintf("incorrect response (first line needs to be SUCCESS): %v", lines)
		}
	} else {
		flags.Error = result
	}
	if err := templ.ExecuteTemplate(w, "flags.html", flags); err != nil {
		fmt.Fprintf(w, "Executing flags template: %v", err)
		return
	}
}

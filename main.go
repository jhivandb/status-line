package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/jhivandb/status-line/internal/api"
	"github.com/jhivandb/status-line/internal/sections"
)

func main() {

	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}

	var inputData api.InputData
	if err := json.Unmarshal(input, &inputData); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	pathSection := &sections.Path{In: inputData}
	gitSection := &sections.Git{In: inputData}
	contextSection := &sections.Context{In: inputData}

	sectionList := []api.Section{pathSection, gitSection, contextSection}

	output := api.ColorReset
	for _, section := range sectionList {
		output += section.Render() + api.ColorReset + " "
	}

	fmt.Println(output)
}

package main

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

var expectedCommands = []string{"kubectl get ns", "echo \"Hello World!\"\nkubectl get pods", "kubectl get ns"}

var mdFile = []string{
	"# Hello World",
	"<!-- command -->",
	"```kubectl get ns```",
	"```kubectl get pods```",
	"<!-- var DT_TENANT -->",
	"<!-- TODO: Add support for code blocks with language -->",
	"<!-- bash\necho \"Hello World!\"\nkubectl get pods\n-->",
	"<!-- bash\nkubectl get ns\n-->",
}

var bashOutputFile = []string{
	"#!/bin/bash",
	"set -e",
	`if [ -z "$DT_TENANT" ]; then`,
	`echo "Please supply a value for the environment variable DT_TENANT"`,
	`exit 1`,
	`fi`,
	"kubectl get ns",
	`echo "Hello World!"`,
	"kubectl get pods",
	"kubectl get ns",
}

func TestGetCommandsAndActionsWithBashInput(t *testing.T) {
	markdownFile := fmt.Sprintln(mdFile)

	htmlFile, err := convertMarkdownToHTML([]byte(markdownFile))

	require.NoError(t, err, "Expected no error")

	commands, err := getCommandsAndActions(htmlFile)

	require.NoError(t, err, "Expected no error")
	require.Equal(t, expectedCommands, commands)
}

func TestProcessComment(t *testing.T) {
	require := require.New(t)

	for _, tt := range []struct {
		Input  string
		Result string
	}{
		{Input: "<!-- bash echo test -->", Result: "echo test"},
		{Input: "<!-- bash kubectl get pods -->", Result: "kubectl get pods"},
		{Input: "<!-- bash wait_for_deployment_with_image_in_namespace -->", Result: "wait_for_deployment_with_image_in_namespace"},
		{Input: "<!-- var DT_TENANT -->", Result: ""},
		{Input: "<!-- var DT_API_TOKEN -->", Result: ""},
	} {
		param := processComment(tt.Input)

		require.Equal(tt.Result, param)
	}
}

func TestConvertTutorialIntoBashFile(t *testing.T) {
	require := require.New(t)

	// Creating markdown file for test
	markdownFile := fmt.Sprintln(mdFile)
	filepath := "testfile.md"
	outputFile := "testoutput.sh"

	deleteFile, err := createAndDeleteMarkdownFile(filepath, markdownFile)

	if deleteFile != nil {
		defer deleteFile()
	}

	require.NoError(err, fmt.Sprintf("Expected no error but got %s", err))

	// Converting markdown file into bash script
	err = convertTutorialIntoBashScript(filepath, outputFile)

	require.NoError(err, fmt.Sprintf("Expected no error but got %s", err))

	// Reading the output of the bash file
	bashFile, err := ioutil.ReadFile(outputFile)

	require.NoError(err, fmt.Sprintf("Expected no error but got %s", err))

	for _, item := range bashOutputFile {
		require.Contains(string(bashFile), item, fmt.Sprintf("Bash file: %s doesn't contain output item %s", string(bashFile), item))
	}

	// Remove output file
	os.Remove(outputFile)
}

func createAndDeleteMarkdownFile(path, content string) (func(), error) {
	err := ioutil.WriteFile(path, []byte(content), 0644)

	if err != nil {
		return nil, err
	}

	return func() {
		os.Remove(path)
	}, nil
}

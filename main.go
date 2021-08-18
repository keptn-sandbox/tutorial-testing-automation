package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	goldHtml "github.com/yuin/goldmark/renderer/html"
	"html"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var command bool
var debug bool
var variables []string

var tagsToRead = []string{
	"code",
	"#comment",
	"h1",
	"h2",
	"h3",
}

func main() {
	var fileName string
	var outputFileName string

	flag.StringVar(&fileName, "f", "", "Markdown file to parse.")
	flag.StringVar(&outputFileName, "o", "", "Name of the bash output file.")

	flag.Parse()

	if fileName == "" {
		fmt.Println("Please provide md file by using -f option")
		return
	}

	if outputFileName == "" {
		fmt.Println("Using default output destination")
		outputFileName = "output.sh"
	}

	if err := convertTutorialIntoBashScript(fileName, outputFileName); err != nil {
		log.Fatal(err)
	}
}

func convertTutorialIntoBashScript(fileName, outputFileName string) error {
	// read content from input file
	htmlBytes, err := readInputFile(fileName)
	if err != nil {
		return err
	}

	// read commands from html
	commands, err := getCommandsAndActions(htmlBytes)
	if err != nil {
		return err
	}

	// generate bash script using the commands
	if err := generateBashScript(commands, outputFileName); err != nil {
		return err
	}

	return nil
}

func readInputFile(fileName string) ([]byte, error) {
	var htmlBytes []byte

	if strings.Split(fileName, ".")[1] == "md" {
		mdFile, err := ioutil.ReadFile(fileName)
		if err != nil {
			return nil, err
		}

		htmlBytes, err = convertMarkdownToHTML(mdFile)

		if err != nil {
			return nil, err
		}
	} else if strings.Split(fileName, ".")[1] == "html" {
		htmlFile, err := ioutil.ReadFile(fileName)
		if err != nil {
			return nil, err
		}

		htmlBytes = htmlFile
	} else {
		log.Fatal("File format not supported. Try again using either and md or html file!")
	}

	return htmlBytes, nil
}

func convertMarkdownToHTML(markdown []byte) ([]byte, error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(),
		goldmark.WithRendererOptions(
			goldHtml.WithUnsafe(),
		),
	)

	var buf bytes.Buffer

	if err := md.Convert(markdown, &buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func getCommandsAndActions(mdFile []byte) ([]string, error) {
	byteReader := bytes.NewReader(mdFile)
	doc, err := goquery.NewDocumentFromReader(byteReader)
	if err != nil {
		return nil, err
	}

	sel := new(goquery.Selection)
	tbl := doc.Find("body")

	// find all wanted items
	sel = visitNodes(sel, tbl)

	// process the data
	commands := processData(sel)

	return commands, nil
}

func processData(sel *goquery.Selection) []string {
	var commands []string

	// loop through items and process the data
	sel.Each(func(i int, s *goquery.Selection) {
		// get the content of the selector
		htmlString, err := goquery.OuterHtml(s)
		if err != nil {
			log.Fatal(err)
		}

		p := bluemonday.StrictPolicy()

		var command string
		switch goquery.NodeName(s) {
		case "#comment":
			htmlString = html.UnescapeString(htmlString)
			command = processComment(htmlString)
		case "code":
			htmlString = p.Sanitize(htmlString)
			htmlString = html.UnescapeString(htmlString)

			command = processCode(htmlString)
		case "h1", "h2", "h3":
			htmlString = p.Sanitize(htmlString)
			htmlString = html.UnescapeString(htmlString)

			printHeading := fmt.Sprintf(`echo "%s"`, htmlString)

			headingCommands := []string{
				`echo "---------------------------------------------------------------------"`,
				printHeading,
				`echo "---------------------------------------------------------------------"`,
				`echo ""`,
			}

			command = strings.Join(headingCommands, "\n")
		default:
			fmt.Println("Node name not supported!")
		}

		if len(command) != 0 {
			commands = append(commands, command)
		}
	})

	return commands
}

func processComment(comment string) string {
	comment = strings.Replace(comment, "<!--", "", -1)
	comment = strings.Replace(comment, "-->", "", -1)
	comment = strings.TrimSpace(comment)

	// Read the different annotations and then process the comment
	if comment == "command" {
		command = true
	} else if strings.Contains(comment, "bash") {
		comment = strings.Replace(comment, "bash", "", -1)
		comment = strings.TrimSpace(comment)
		return comment
	} else if comment == "debug" {
		debug = true
		command = true
	} else if strings.Contains(comment, "var") {
		comment = strings.Replace(comment, "var", "", -1)
		comment = strings.TrimSpace(comment)

		comment = fmt.Sprintf("if [ -z \"$%s\" ]; then\n 	echo \"Please supply a value for the environment variable %s\"\n	exit 1\nfi", comment, comment)

		variables = append(variables, comment)
		return ""
	}

	return ""
}

func processCode(code string) string {
	code = strings.TrimSpace(code)

	if command && debug {
		command = false
		debug = false
		return fmt.Sprintf("if [ \"$DEBUG\" = \"true\" ]; then \n%s  \nfi", code)
	} else if command {
		command = false
		return code
	}

	return ""
}

func visitNodes(dst, src *goquery.Selection) *goquery.Selection {
	src.Contents().Each(func(i int, s *goquery.Selection) {
		if sliceContains(tagsToRead, goquery.NodeName(s)) {
			dst = dst.AddSelection(s)
		} else {
			dst = visitNodes(dst, s)
		}
	})

	return dst
}

func generateBashScript(commands []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	importUtilsPath := `source /dev/stdin <<<"$( curl -sS https://raw.githubusercontent.com/keptn/keptn/master/test/utils.sh)"`

	fmt.Fprintln(w, "#!/bin/bash")
	fmt.Fprintln(w, "set -e")
	fmt.Fprintln(w, importUtilsPath)
	fmt.Fprintln(w, "")

	for _, line := range variables {
		fmt.Fprintln(w, line)
		fmt.Fprintln(w, "")
	}

	for _, line := range commands {
		fmt.Fprintln(w, line)
		fmt.Fprintln(w, "")
	}

	err = os.Chmod(path, 0777)
	if err != nil {
		return err
	}

	return w.Flush()
}

func sliceContains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
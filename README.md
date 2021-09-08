# Keptn tutorials run automation

This tool allows you to automatically run Keptn tutorials by converting annotated Markdown files into a bash file.

## Table of content

- [Prerequisites](#prerequisites)
- [Installing the tool](#installing-the-tool)
- [Converting Markdown file to bash](#converting-markdown-file-to-bash)
- [Available annotations](#available-annotations)

## Prerequisites

Please download and install the following tools before continuing:

- [Go](https://golang.org/dl/)

## Installing the tool

Download the tool by cloning the repository from GitHub.

You can then install the tool by moving into the directory and executing the following command:

```bash
go install github.com/keptn-sandbox/tutorial-testing-automation
```

You should now be able to execute the tool by typing `tutorial-testing-automation` in your command line. If this doesn't work you can still use the application by running `go run main.go` instead of `tutorial-testing-automation`.

If you just want to run the tool you can also just install it through Go:

```bash
go get github.com/keptn-sandbox/tutorial-testing-automation
```

## Converting Markdown file to bash

Annotated Markdown files can easily be converted to bash using the following command:

```bash
tutorial-testing-automation -f FILENAME.md
```

## Available annotations

The following annotations allow you to define which commands should be copied from your markdown file into the bash script. They also allow you to define extra verification steps, commands and variables.

### Command annotation

The `command` annotation can be used together with a code block below to signal that the code block should be copied into the bash script. If this annotation is not defined the code block will be ignored.

#### Example

For the following markdown input:

~~~
<!-- command -->
```
keptn create project unleash --shipyard=./shipyard.yaml
```

```
echo http://unleash.unleash-dev.$(kubectl -n keptn get ingress api-keptn-ingress -ojsonpath='{.spec.rules[0].host}')
```
~~~

You will get the following bash output:

```bash
keptn create project unleash --shipyard=./shipyard.yaml
```

As you can see the code block without the `command` annotation will not be copied into the generated bash file.

### Bash annotation

The `bash` annotation lets you define extra commands or call functions from the bash utilities that will not be visible in the tutorial but still need to be executed in the automation script.

A perfect example of this is changing directory which is normally only written in text and executed by the user.

```md
<!-- bash cd ../.. -->
```

You can also call functions defined in the `utils.sh` file of the repository.

```md
<!-- bash
verify_test_step $? "Send event new-artifact for unleash failed"
wait_for_deployment_with_image_in_namespace "unleash-db" "unleash-dev" "postgres:10.4"
wait_for_deployment_with_image_in_namespace "unleash" "unleash-dev" "docker.io/keptnexamples/unleash:1.0.0"
-->
```

### Var annotation

The `var` annotation lets you define variables that will be used by the script and need to be set before the functionality can be run.

If your tutorial for example needs a Dynatrace tenant:

```md
<!-- var DT_TENANT -->
```

This will leave you with the following bash output at the beginning of the bash file:

```
if [ -z "$DT_TENANT" ]; then
 	echo "Please supply a value for the environment variable DT_TENANT"
	exit 1
fi
```

### Debug annotation

The `debug` annotation works similar to the `command` one with the slight difference that is puts the code of the annotated code block into an if statement that is only executed when the `DEBUG` environment variable is set to true.

~~~
<!-- debug -->
```
kubectl get pods -n dynatrace
```
~~~

Output:

```bash
if [ "$DEBUG" = "true" ]; then kubectl get pods -n dynatrace ; fi
```

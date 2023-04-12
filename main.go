package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"
)

const (
	defaultRegistry = "ghcr.io"
)

var (
	branch     = os.Getenv("GITHUB_REF_NAME")
	registry   = os.Getenv("INPUT_REGISTRY")
	username   = os.Getenv("INPUT_USERNAME")
	password   = os.Getenv("INPUT_PASSWORD")
	image      = os.Getenv("INPUT_IMAGE")
	tag        = os.Getenv("INPUT_TAG")
	repository = image
	context    = path.Join(os.Getenv("GITHUB_WORKSPACE"), os.Getenv("INPUT_PATH"))
	dockerfile = os.Getenv("INPUT_DOCKERFILE")
	tagLatest  = os.Getenv("INPUT_TAG_WITH_LATEST") == "true"
	target     = os.Getenv("INPUT_TARGET")

	imageLatest = ""

	cache    = os.Getenv("INPUT_CACHE") == "true"
	cacheTTL = os.Getenv("INPUT_CACHE_TTL")
	cacheURL = os.Getenv("INPUT_CACHE_URL")

	buildArgs = strings.Split(os.Getenv("INPUT_BUILD_ARGS"), ",")
)

func init() {
	if tag == "" {
		if branch == "dev" || branch == "main" || branch == "master" {
			tag = "latest"
		}
	}

	if dockerfile == "" {
		dockerfile = "Dockerfile"
	}

	if tagLatest {
		imageLatest = fmt.Sprintf("%s:%s", repository, "latest")
	}

	if registry == "" {
		registry = defaultRegistry
	}

	if username == "" {
		username = os.Getenv("GITHUB_ACTOR")
	}
	if password == "" {
		password = os.Getenv("GITHUB_TOKEN")
	}

	if cache && cacheTTL == "" {
		cacheTTL = "48h"
	}

	if registry == "ghcr.io" {
		namespace := strings.ToLower(os.Getenv("GITHUB_REPOSITORY"))
		image = namespace + "/" + image
		repository = namespace + "/" + repository

		if tagLatest {
			imageLatest = namespace + "/" + imageLatest
		}

	}

	if registry == "docker.io" {
		registry = fmt.Sprintf("index.%s/v1/", registry)
	}

	if cacheURL == "" {
		cacheURL = registry + "/" + image + ":cache"
	}
}

func scanImages() []string {
	// tags arrive as lines of text delimited by newline \n
	tgs := os.Getenv("INPUT_IMAGES")
	imgInput := bytes.NewReader([]byte(tgs))
	scanner := bufio.NewScanner(imgInput)

	var images []string
	for scanner.Scan() {
		images = append(images, scanner.Text())
	}

	return images
}

var dockerConfigJson = `
{
  "auths": {
	"https://{{.Registry}}": {
	  "username": "{{.Username}}",
	  "password": "{{.Password}}"
	}
  }
}
`

type auth struct {
	Registry string `json:"registry"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func writeAuth() {
	tpl, err := template.New("dockerlogin").Parse(dockerConfigJson)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile("/kaniko/.docker/config.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	auth := &auth{
		Registry: registry,
		Username: username,
		Password: password,
	}

	err = tpl.Execute(f, auth)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	cmdArgs := make([]string, 0)

	// reproducible is still broken since 1.7.0 :( - https://github.com/GoogleContainerTools/kaniko/issues/2005
	// cmdArgs = append(cmdArgs, "--reproducible", "--force", "--verbosity=info")
	cmdArgs = append(cmdArgs, "--force", "--verbosity=info")

	if cache {
		cmdArgs = append(cmdArgs, "--cache=true", "--cache-ttl=48h", fmt.Sprintf("--cache-repo=%s", cacheURL))
	}

	if target != "" {
		cmdArgs = append(cmdArgs, "--target", target)
	}

	if len(buildArgs) > 0 {
		for _, arg := range buildArgs {
			cmdArgs = append(cmdArgs, fmt.Sprintf("--build-arg=%s", arg))
		}
	}

	cmdArgs = append(cmdArgs, "--context", context)
	cmdArgs = append(cmdArgs, "--dockerfile", dockerfile)
	cmdArgs = append(cmdArgs, "--snapshotMode=redo")

	if image == "" {
		images := scanImages()
		if len(images) == 0 {
			output(fmt.Errorf("no image name provided from either 'image' or 'images' inputs"))
			os.Exit(1)
		}
		for _, img := range images {
			cmdArgs = append(cmdArgs, "--destination", registry+"/"+img)
		}
	} else {
		cmdArgs = append(cmdArgs, "--destination", fmt.Sprintf("%s/%s:%s", registry, image, tag))
	}

	writeAuth()

	fmt.Println(cmdArgs)

	cmd := exec.Command("/kaniko/executor", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		output(fmt.Errorf("error starting Cmd: %w", err))
		os.Exit(1)
	}

	if err := cmd.Wait(); err != nil {
		output(err)
		os.Exit(1)
	}
}

func output(err error) {
	fmt.Println(fmt.Printf("::warning ::%s", err.Error()))
}

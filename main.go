package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	yaml "gopkg.in/yaml.v2"
)

type miniDeployment struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string
	Metadata   struct {
		Name string
	}
	Spec struct {
		Template struct {
			Spec struct {
				Containers []struct {
					Name  string
					Image string
				}
			}
		}
	}
}

func patchFile(filename string, oldTag, newTag string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to patch %s - skipping\n", filename)
		return
	}
	var reImage = regexp.MustCompile(fmt.Sprintf("image:\\s*['\"]*%s\\s*['\"]*\n", oldTag))
	newData := reImage.ReplaceAllLiteral(data, []byte("image: "+newTag+"\n"))
	ioutil.WriteFile(filename, newData, 0644)
}

func investigateFile(t miniDeployment, filename string, image string, newTag string) {
	if t.Kind == "Deployment" {
		for _, container := range t.Spec.Template.Spec.Containers {
			parts := strings.SplitN(container.Image, ":", 2)
			if parts[0] == image {
				baseName := path.Base(filename)
				if container.Image != newTag {
					fmt.Fprintf(os.Stderr, "%s (%s/%s) %s -> %s\n", baseName, t.Metadata.Name, container.Name, container.Image, newTag)
					patchFile(filename, container.Image, newTag)
				}
			}
		}
	}
}

func getLatestTag() string {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		panic(err)
	}

	for _, image := range images {
		if len(image.RepoTags) == 0 || image.RepoTags[0] == "<none>:<none>" {
			continue
		}
		return image.RepoTags[0]
	}
	return ""
}

func main() {
	templatePath := os.Getenv("UPDATE_DOCKER_TAG_PATH")
	if templatePath == "" {
		templatePath = "."
	}

	var newTag string
	if len(os.Args) >= 2 {
		newTag = os.Args[1]
	} else {
		newTag = getLatestTag()
	}

	image := strings.SplitN(newTag, ":", 2)[0]

	err := filepath.Walk(templatePath, func(path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".yaml") {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read file %s - skipping.\n", path)
			} else {
				t := miniDeployment{}
				err = yaml.Unmarshal(data, &t)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to parse YAML in %s - skipping.\n", path)
				} else {
					investigateFile(t, path, image, newTag)
				}
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

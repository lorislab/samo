package helm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lorislab/samo/file"
	"github.com/lorislab/samo/project"
	log "github.com/sirupsen/logrus"
)

type HelmRequest struct {
	Input    string
	Output   string
	Clean    bool
	Template string
}

var (
	regexProjectName    = regexp.MustCompile(`\$\{project\.name\}`)
	regexProjectVersion = regexp.MustCompile(`\$\{project\.version\}`)
)

// Filter
func (request HelmRequest) Filter(p project.Project) {

	template := request.Template
	if len(template) == 0 {
		log.WithField("template", request.Template).Warn("Template is empty! Switch back to maven template")
		template = "maven"
	}
	templateF := templateFunc(template)

	// output directory output + project.name
	outputDir := filepath.FromSlash(request.Output + "/" + p.Name())

	if _, err := os.Stat(request.Input); os.IsNotExist(err) {
		log.WithField("input", request.Input).Fatal("Input helm directory does not exists!")
	}

	// clean output directory
	if request.Clean {
		if _, err := os.Stat(request.Output); !os.IsNotExist(err) {
			err := os.RemoveAll(request.Output)
			if err != nil {
				log.WithField("output", request.Output).Panic(err)
			}
		}
	}

	// get all files from the input directory
	paths, err := file.GetAllFilePathsInDirectory(request.Input)
	if err != nil {
		log.WithField("input", request.Input).Panic(err)
	}

	for _, path := range paths {
		// load file
		result, err := ioutil.ReadFile(path)
		if err != nil {
			log.Panic(err)
		}
		result = templateF(p, result)
		// write result to output directory
		file.WriteBytesToFile(strings.Replace(path, request.Input, outputDir, -1), result)
	}

}

func templateFunc(template string) func(p project.Project, data []byte) []byte {
	switch template {
	case "maven":
		return func(p project.Project, data []byte) []byte {
			return templateMavenFilter(p, data)
		}
	default:
		log.WithField("template", template).Fatal("Not supported template!")
	}
	return nil
}

func templateMavenFilter(p project.Project, data []byte) []byte {
	result := regexProjectName.ReplaceAll(data, []byte(p.Name()))
	result = regexProjectVersion.ReplaceAll(result, []byte(p.Version()))
	return result
}

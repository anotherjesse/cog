package api

import (
	"fmt"
	"log"
	"os"

	"github.com/replicate/cog/pkg/docker"
	"github.com/replicate/cog/pkg/image"
	"github.com/replicate/cog/pkg/predict"
	"github.com/replicate/cog/pkg/util/console"
)

type Engine struct {
	version   string
	versionDB map[string]Version
	p         *predict.Predictor
	result    *Response
}

func NewEngine() *Engine {
	return &Engine{
		versionDB: map[string]Version{},
	}
}

func (e *Engine) AddVersion(v Version) {
	e.versionDB[v.versionID] = v
}
func (e *Engine) HasVersion(versionID string) bool {
	_, ok := e.versionDB[versionID]
	return ok
}
func (e *Engine) GetVersion(versionID string) *Version {
	if v, ok := e.versionDB[versionID]; ok {
		return &v
	}
	return nil
}

func (e *Engine) Predict(body []byte) {
	if e.p == nil {
		fmt.Println("No model loaded")
		return
	}

	if e.result == nil {
		fmt.Println("No result")
		return
	}

	response, err := e.p.PredictJSON(body)
	if err != nil {
		log.Println("error predicting", err)
		e.result.Status = "failed"
		return
	}

	e.result.Output = response.Output
	e.result.Status = "succeeded"
}

func (e *Engine) LoadVersion(imageName string, version string) error {

	// if already loaded, do nothing
	if e.version == version {
		return nil
	}

	if err := ensureImageExists(imageName); err != nil {
		return err
	}

	conf, err := image.GetConfig(imageName)
	if err != nil {
		return err
	}

	gpus := ""
	volumes := []docker.Volume{}
	if conf.Build.GPU {
		gpus = "all"
	}

	if e.p != nil {
		if err := e.p.Stop(); err != nil {
			console.Warnf("Failed to stop container: %s", err)
		}
	}

	fmt.Println("predictor: loading version", version)
	p := predict.NewPredictor(docker.RunOptions{
		GPUs:    gpus,
		Image:   imageName,
		Volumes: volumes,
	})

	if err := p.Start(os.Stderr); err != nil {
		return err
	}

	e.p = &p
	e.version = version

	fmt.Println("predictor: ready version", version)

	return nil
}

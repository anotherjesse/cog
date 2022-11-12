package api

import (
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
		console.Info("No model loaded")
		return
	}

	if e.result == nil {
		console.Infof("No result")
		return
	}

	response, err := e.p.PredictJSON(body)
	if err != nil {
		console.Warnf("error predicting: %s", err)
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
		console.Infof("Stopping container for model version %s", e.version)
		if err := e.p.Stop(); err != nil {
			console.Warnf("Failed to stop container: %s", err)
		}
	}

	console.Infof("Loading model version %s", version)

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

	console.Infof("Ready model version %s", version)

	return nil
}

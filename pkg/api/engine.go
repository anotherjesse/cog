package api

import (
	"errors"
	"os"

	"github.com/replicate/cog/pkg/docker"
	"github.com/replicate/cog/pkg/image"
	"github.com/replicate/cog/pkg/predict"
	"github.com/replicate/cog/pkg/util/console"
)

type Engine struct {
	currentVersion string
	versionDB      map[string]Version
	p              *predict.Predictor
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

func (e *Engine) Predict(r *Response) error {

	body, err := LoadPredictionInput(r.ID)
	if err != nil {
		console.Warnf("unable to read request body: %s", err)
		return err
	}
	if e.p == nil {
		return errors.New("predictor not loaded")
	}

	prediction, err := e.p.PredictJSON(body)
	if err != nil {
		console.Warnf("error predicting: %s", err)
		r.Status = "failed"
		r.Save()
		return nil
	}

	r.Status = "succeeded"
	r.Output = prediction.Output
	if err := r.Save(); err != nil {
		console.Warnf("error saving prediction: %s", err)
		return err
	}

	return nil
}

func (e *Engine) LoadVersion(imageName string, version string) error {

	// if already loaded, do nothing
	if e.currentVersion == version {
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
		console.Infof("Stopping container for model version %s", e.currentVersion)
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
	e.currentVersion = version

	console.Infof("Ready model version %s", version)

	return nil
}

func (e *Engine) Run(queue chan string) {
	for {
		predictionID := <-queue
		console.Infof("Running prediction %s", predictionID)
		p, err := LoadPrediction(predictionID)

		if err != nil {
			console.Warnf("error loading prediction: %s", err)
			continue
		}

		v := e.GetVersion(p.Version)
		if v == nil {
			console.Warnf("version not found: %s", p.Version)
			console.Warnf("this is only populated if the openapi spec is requested :(")
			continue
		}

		if err := e.LoadVersion(v.imageName(), p.Version); err != nil {
			console.Warnf("unable to load version: %s", err)
			continue
		}

		err = e.Predict(p)
		if err != nil {
			console.Warnf("predict runner errored: %s", err)
			continue
		}
	}

}

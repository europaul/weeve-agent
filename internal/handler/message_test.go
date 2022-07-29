package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/Jeffail/gabs/v2"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/weeveiot/weeve-agent/internal/dataservice"
	"github.com/weeveiot/weeve-agent/internal/docker"
	"github.com/weeveiot/weeve-agent/internal/handler"
	"github.com/weeveiot/weeve-agent/internal/manifest"
	"github.com/weeveiot/weeve-agent/internal/model"
)

func init() {
	docker.SetupDockerClient()
}

var manCmd struct {
	ManifestName  string  `json:"manifestName"`
	VersionNumber float64 `json:"versionNumber"`
	Command       string  `json:"command"`
}

var ctx = context.Background()
var dockerCli *client.Client

func TestProcessMessagePass(t *testing.T) {
	assert := assert.New(t)
	// Prepare test data
	manifestPath := "../../testdata/test_manifest.json"
	jsonBytes, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}

	jsonParsed, err := gabs.ParseJSON(jsonBytes)
	if err != nil {
		t.Fatal(err)
	}

	man, err := parseManifest(jsonParsed)
	if err != nil {
		t.Fatal(err)
	}

	manCmd.ManifestName = man.ManifestUniqueID.ManifestName
	manCmd.VersionNumber = man.VersionNumber

	dockerCli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("TESTING EDGE APPLICATION DEPLOYMENT...")
	err = deployEdgeApplication(jsonBytes, man)
	if err != nil {
		err = undeployEdgeApplication(man, dataservice.CMDRemove)
		if err != nil {
			t.Fatal(err)
		}
		t.FailNow()
	}

	fmt.Println("TESTING STOP EDGE APPLICATION...")
	err = stopEdgeApplication(man)
	if err != nil {
		err = undeployEdgeApplication(man, dataservice.CMDRemove)
		if err != nil {
			t.Fatal(err)
		}
		t.FailNow()
	}

	fmt.Println("TESTING START EDGE APPLICATION...")
	err = startEdgeApplication(man)
	if err != nil {
		err = undeployEdgeApplication(man, dataservice.CMDRemove)
		if err != nil {
			t.Fatal(err)
		}
		t.FailNow()
	}

	fmt.Println("TESTING REDEPLOY EDGE APPLICATION...")
	err = reDeployEdgeApplication(jsonBytes, man)
	if err != nil {
		err = undeployEdgeApplication(man, dataservice.CMDRemove)
		if err != nil {
			t.Fatal(err)
		}
		t.FailNow()
	}

	fmt.Println("TESTING UNDEPLOY EDGE APPLICATION...")
	err = undeployEdgeApplication(man, dataservice.CMDUndeploy)
	if err != nil {
		err = undeployEdgeApplication(man, dataservice.CMDRemove)
		if err != nil {
			t.Fatal(err)
		}
		t.FailNow()
	}

	fmt.Println("DEPLOYING EDGE APPLICATION FOR TESTING REMOVE EDGE APPLICATION...")
	err = deployEdgeApplication(jsonBytes, man)
	if err != nil {
		err = undeployEdgeApplication(man, dataservice.CMDRemove)
		if err != nil {
			t.Fatal(err)
		}
		t.FailNow()
	}

	fmt.Println("TESTING REMOVE EDGE APPLICATION...")
	err = undeployEdgeApplication(man, dataservice.CMDRemove)
	assert.Nil(err)
}

func deployEdgeApplication(jsonBytes []byte, man manifest.Manifest) error {
	// Process deploy edge application
	err := handler.ProcessMessage(jsonBytes)
	if err != nil {
		return fmt.Errorf("ProcessMessage returned %v status", err)
	}

	// Verify deployment
	net, err := getNetwork(man.ManifestUniqueID)
	if err != nil {
		return err
	}

	if len(net) > 0 {
		_, err := checkContainersExistsWithStatus(man.ManifestUniqueID, len(man.Modules), "running")
		if err != nil {
			return err
		}
	} else {
		return errors.New("Network not created")
	}

	return nil
}

func reDeployEdgeApplication(jsonBytes []byte, man manifest.Manifest) error {
	jsonParsed, err := gabs.ParseJSON(jsonBytes)
	if err != nil {
		return err
	}

	jsonParsed.Set("REDEPLOY", "command")
	jsonBytes = jsonParsed.Bytes()

	currentTime := time.Now()
	// Process deploy edge application
	err = handler.ProcessMessage(jsonBytes)
	if err != nil {
		return fmt.Errorf("ProcessMessage returned %v status", err)
	}

	net, err := getNetwork(man.ManifestUniqueID)
	if err != nil {
		return err
	}

	if len(net) > 0 {
		if net[0].Created.After(currentTime) {
			_, err := checkContainersExistsWithStatus(man.ManifestUniqueID, len(man.Modules), "running")
			if err != nil {
				return err
			}
		} else {
			return errors.New("Expected new network, found old network")
		}
	} else {
		return errors.New("Network not created")
	}

	return nil
}

func stopEdgeApplication(man manifest.Manifest) error {
	// Process stop edge application
	manCmd.Command = dataservice.CMDStopService
	jsonB, err := json.Marshal(manCmd)
	if err != nil {
		panic(err)
	}

	err = handler.ProcessMessage(jsonB)
	if err != nil {
		return fmt.Errorf("ProcessMessage returned %v status", err)
	}

	_, err = checkContainersExistsWithStatus(man.ManifestUniqueID, len(man.Modules), "exited")
	if err != nil {
		return err
	}

	return nil
}

func startEdgeApplication(man manifest.Manifest) error {
	// Process start edge application
	manCmd.Command = dataservice.CMDStartService
	jsonB, err := json.Marshal(manCmd)
	if err != nil {
		return err
	}

	err = handler.ProcessMessage(jsonB)
	if err != nil {
		return fmt.Errorf("ProcessMessage returned %v status", err)
	}

	_, err = checkContainersExistsWithStatus(man.ManifestUniqueID, len(man.Modules), "running")
	if err != nil {
		return err
	}

	return nil
}

func undeployEdgeApplication(man manifest.Manifest, operation string) error {
	if operation == dataservice.CMDUndeploy || operation == dataservice.CMDRemove {
		manCmd.Command = operation
	} else {
		return errors.New("Invalid operation: " + operation)
	}
	jsonB, err := json.Marshal(manCmd)
	if err != nil {
		return err
	}

	err = handler.ProcessMessage(jsonB)
	if err != nil {
		return fmt.Errorf("ProcessMessage returned %v status", err)
	}

	net, err := getNetwork(man.ManifestUniqueID)
	if err != nil {
		return err
	}

	if len(net) <= 0 {
		dsContainers, _ := getEdgeApplicationContainers(man.ManifestUniqueID)
		if len(dsContainers) > 0 {
			return errors.New("Edge application undeployment failed, containers not deleted")
		}
	} else {
		return errors.New("Edge application undeployment failed, network not deleted")
	}

	if operation == dataservice.CMDUndeploy {
		exist, err := checkImages(man, true)
		if err != nil {
			return err
		}

		if !exist {
			return errors.New("Edge application undeploy should not delete images")
		}
	} else {
		deleted, err := checkImages(man, false)
		if err != nil {
			return err
		}

		if !deleted {
			return errors.New("Edge application removal should delete images")
		}
	}

	return nil
}

func parseManifest(jsonParsed *gabs.Container) (manifest.Manifest, error) {
	manifestName := jsonParsed.Search("manifestName").Data().(string)
	versionNumber := jsonParsed.Search("versionNumber").Data().(float64)

	var containerConfigs []manifest.ContainerConfig

	modules := jsonParsed.Search("modules").Children()
	for _, module := range modules {
		var containerConfig manifest.ContainerConfig

		containerConfig.ImageName = module.Search("image").Search("name").Data().(string)
		containerConfig.ImageTag = module.Search("image").Search("tag").Data().(string)

		imageName := containerConfig.ImageName
		if containerConfig.ImageTag != "" {
			imageName = imageName + ":" + containerConfig.ImageTag
		}

		containerConfig.Registry = manifest.RegistryDetails{ImageName: imageName}
		containerConfigs = append(containerConfigs, containerConfig)
	}

	manifest := manifest.Manifest{
		ManifestUniqueID: model.ManifestUniqueID{ManifestName: manifestName, VersionNumber: fmt.Sprint(versionNumber)},
		VersionNumber:    versionNumber,
		Modules:          containerConfigs,
	}

	return manifest, nil
}

func getNetwork(manID model.ManifestUniqueID) ([]types.NetworkResource, error) {
	filter := filters.NewArgs()
	filter.Add("label", "manifestName="+manID.ManifestName)
	filter.Add("label", "versionNumber="+manID.VersionNumber)
	options := types.NetworkListOptions{Filters: filter}
	networks, err := dockerCli.NetworkList(context.Background(), options)
	if err != nil {
		return nil, err
	}

	return networks, nil
}

func getEdgeApplicationContainers(manifestUniqueID model.ManifestUniqueID) ([]types.Container, error) {
	filter := filters.NewArgs()
	filter.Add("label", "manifestName="+manifestUniqueID.ManifestName)
	filter.Add("label", "versionNumber="+manifestUniqueID.VersionNumber)
	options := types.ContainerListOptions{All: true, Filters: filter}
	containers, err := dockerCli.ContainerList(context.Background(), options)
	if err != nil {
		return nil, err
	}

	return containers, nil
}

func checkContainersExistsWithStatus(manID model.ManifestUniqueID, containerCount int, status string) (bool, error) {
	dsContainers, _ := getEdgeApplicationContainers(manID)
	if containerCount != len(dsContainers) {
		return false, fmt.Errorf("Expected number of containers %v, number of available containers %v", containerCount, len(dsContainers))
	}
	for _, dsContainer := range dsContainers {
		if dsContainer.State != status {
			return false, fmt.Errorf("Container expected status %s, but current status %s", status, dsContainer.State)
		}
	}

	return true, nil
}

func checkImages(man manifest.Manifest, exist bool) (bool, error) {

	for _, module := range man.Modules {
		imgDetails := module.Registry
		_, _, err := dockerCli.ImageInspectWithRaw(ctx, imgDetails.ImageName)
		if err != nil {
			if client.IsErrNotFound(err) {
				if exist {
					return false, nil
				}
			} else {
				return false, err
			}
		} else if !exist {
			return false, err
		}
	}

	return true, nil
}

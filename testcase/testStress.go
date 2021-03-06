package testcase

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/kunalkushwaha/ctr-powertest/libruntime"
	log "github.com/sirupsen/logrus"
)

type StressTest struct {
	Runtime libruntime.Runtime
}

var testCaseMap map[string]bool

func (t *StressTest) RunTestCases(ctx context.Context, testcases, args []string) error {
	log.Info("Running tests on ", t.Runtime.Version(ctx))

	for _, test := range testcases {
		switch test {
		case "container-create-delete":
			if err := t.TestContainerCreateDelete(ctx, 4, 50); err != nil {
				return err
			}
		case "image-pull":
			var imageName string
			if len(args) == 0 {
				imageName = "docker.io/library/golang:latest"
			} else {
				imageName = string(args[0])
			}

			if err := t.TestImagePull(ctx, 4, imageName); err != nil {
				return err
			}
		}
	}

	if len(testcases) == 0 {

		if err := t.TestContainerCreateDelete(ctx, 4, 50); err != nil {
			return err
		}
		if err := t.TestImagePull(ctx, 4, "docker.io/library/golang:latest"); err != nil {
			return err
		}
	}

	return nil
}

func (t *StressTest) TestContainerCreateDelete(ctx context.Context, parallelCount, loopCount int) error {
	//Make Sure image exist
	_, err := t.Runtime.Pull(ctx, testImage)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(parallelCount)
	log.Infof("Creating %d container in %d goroutines", loopCount, parallelCount)
	startTime := time.Now()
	for i := 0; i < parallelCount; i++ {
		go t.createDeleteContainers(ctx, i, loopCount, &wg)
	}
	wg.Wait()
	totalTime := time.Now().Sub(startTime)
	log.Infof("%d containers in %s ", loopCount*parallelCount, totalTime.String())

	log.Info("OK")
	return nil
}

func (t *StressTest) createDeleteContainers(ctx context.Context, id, loopCount int, wg *sync.WaitGroup) error {
	defer wg.Done()
	//Genertate Specs
	//Seed random number
	//Image name
	for i := 0; i < loopCount; i++ {

		ctr, err := t.Runtime.Create(ctx, testContainerName+"-"+strconv.Itoa(id+1020)+"-"+strconv.Itoa(i+1010), testImage, nil)
		if err != nil {
			log.Error(err)
			return err
		}

		err = t.Runtime.Runnable(ctx, ctr)
		if err != nil {
			log.Error(err)
			return err
		}

		statusC, err := t.Runtime.Wait(ctx, ctr)
		if err != nil {
			log.Error(err)
			return err
		}

		err = t.Runtime.Start(ctx, ctr)
		if err != nil {
			log.Error(err)
			return err
		}

		waitForContainerEvent(statusC)

		err = t.Runtime.Stop(ctx, ctr)
		if err != nil {
			log.Error(err)
			return err
		}

		err = t.Runtime.Delete(ctx, ctr)
		if err != nil {
			log.Error(err)
			return err
		}

	}
	return nil
}

func (t *StressTest) TestImagePull(ctx context.Context, parallelCount int, imageName string) error {
	var wg sync.WaitGroup
	wg.Add(parallelCount)
	log.Infof("Pulling image in %d goroutines", parallelCount)

	for i := 0; i < parallelCount; i++ {
		go t.pullImage(ctx, imageName, &wg)
	}
	wg.Wait()
	t.Runtime.RemoveImage(ctx, imageName)
	log.Info("OK")
	return nil
}

func (t *StressTest) pullImage(ctx context.Context, imageName string, wg *sync.WaitGroup) error {
	defer wg.Done()
	_, err := t.Runtime.Pull(ctx, imageName)
	if err != nil {
		log.Error("Image pull error : ", err)
		return err
	}
	return nil
}

/*
	TODO:

	- Pull & delete images at same time
	- Delete and Exec Containers
*/

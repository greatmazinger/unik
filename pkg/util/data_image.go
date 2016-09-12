// +build !container-binary

package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	unikos "github.com/emc-advanced-dev/unik/pkg/os"
)

func BuildRawDataImage(dataTar io.ReadCloser, size unikos.MegaBytes, usePartitionTables bool) (string, error) {
	dataFolder, err := ioutil.TempDir("", ".raw_data_image_folder.")
	if err != nil {
		return "", errors.New("creating tmp build folder", err)
	}
	defer os.RemoveAll(dataFolder)

	if err := unikos.ExtractTar(dataTar, dataFolder); err != nil {
		return "", errors.New("extracting data tar", err)
	}

	buildDir := filepath.Dir(dataFolder)

	container := NewContainer("image-creator").Privileged(true).WithVolume("/dev/", "/dev/").
		WithVolume(buildDir+"/", "/opt/vol")

	resultFile, err := ioutil.TempFile("", "data.image.result.img.")
	if err != nil {
		return "", err
	}
	resultFile.Close()
	args := []string{"-o", filepath.Base(resultFile.Name())}

	if size > 0 {
		args = append(args, "-p", fmt.Sprintf("%v", usePartitionTables),
			"-v", fmt.Sprintf("%s,%v", filepath.Base(dataFolder), size.ToBytes()))
	} else {
		args = append(args, "-p", fmt.Sprintf("%v", usePartitionTables),
			"-v", filepath.Base(dataFolder),
		)
	}

	logrus.WithFields(logrus.Fields{
		"command": args,
	}).Debugf("running image-creator container")

	if err = container.Run(args...); err != nil {
		return "", errors.New("failed running image-creator on "+dataFolder, err)
	}

	return resultFile.Name(), nil
}

func BuildEmptyDataVolume(size unikos.MegaBytes) (string, error) {
	if size < 1 {
		return "", errors.New("must specify size > 0", nil)
	}
	dataFolder, err := ioutil.TempDir("", "empty.data.folder.")
	if err != nil {
		return "", errors.New("creating tmp build folder", err)
	}
	defer os.RemoveAll(dataFolder)

	buildDir := filepath.Dir(dataFolder)

	container := NewContainer("image-creator").Privileged(true).WithVolume("/dev/", "/dev/").
		WithVolume(buildDir+"/", "/opt/vol")

	resultFile, err := ioutil.TempFile("", "data.image.result.img.")
	if err != nil {
		return "", err
	}
	resultFile.Close()
	args := []string{"-v", fmt.Sprintf("%s,%v", filepath.Base(dataFolder), size.ToBytes()), "-o", filepath.Base(resultFile.Name())}

	logrus.WithFields(logrus.Fields{
		"command": args,
	}).Debugf("running image-creator container")
	if err := container.Run(args...); err != nil {
		return "", errors.New("failed running image-creator on "+dataFolder, err)
	}

	return resultFile.Name(), nil
}

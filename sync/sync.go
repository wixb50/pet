package sync

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/araddon/dateparse"
	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
	"pet/config"
	"pet/snippet"
)

// AutoSync syncs snippets automatically
func AutoSync(file string) error {
	ossMeta, err := getObjectMeta(file)
	if err != nil {
		return err
	}
	lastModified, ok := ossMeta["Last-Modified"]
	if !ok {
		return upload()
	}

	fi, err := os.Stat(file)
	if os.IsNotExist(err) {
		return download()
	} else if err != nil {
		return errors.Wrap(err, "Failed to get a FileInfo")
	}

	local := fi.ModTime().UTC()
	remote, err := dateparse.ParseLocal(lastModified[0])
	if err != nil {
		return errors.Wrap(err, "Failed to get a FileInfo")
	}

	switch {
	case local.After(remote):
		return upload()
	case remote.After(local):
		return download()
	default:
		return nil
	}
}

func getObjectMeta(file string) (http.Header, error) {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Start()
	s.Suffix = " Getting Object..."
	defer s.Stop()

	var snippetName = filepath.Base(file)
	client := aliOSSClient()
	props, err := client.GetObjectDetailedMeta(snippetName)
	if err != nil {
		if len(props) == 0 {
			return props, nil
		}
		return nil, errors.Wrapf(err, "Failed to get object")
	}
	return props, nil
}

func upload() (err error) {
	var snippetFile = config.Conf.General.SnippetFile
	var snippetName = filepath.Base(snippetFile)

	var snippets snippet.Snippets
	if err := snippets.Load(); err != nil {
		return err
	}

	body, err := snippets.ToString()
	if err != nil {
		return err
	}
	if len(body) == 0 {
		fmt.Printf("%s is empty, skip sync\n", snippetName)
		return nil
	}

	client := aliOSSClient()
	err = client.PutObjectFromFile(snippetFile, snippetName)
	if err != nil {
		return errors.Wrapf(err, "%s upload error", snippetName)
	}
	fmt.Println("Upload success")
	return nil
}

func download() error {
	var snippetFile = config.Conf.General.SnippetFile
	var snippetName = filepath.Base(snippetFile)

	client := aliOSSClient()
	err := client.GetObjectToFile(snippetFile, snippetFile)
	if err != nil {
		return errors.Wrapf(err, "%s download error", snippetName)
	}

	fmt.Println("Already up-to-date")
	return nil
}

func aliOSSClient() *oss.Bucket {
	client, _ := oss.New(
		config.Conf.AliOSS.Endpoint, config.Conf.AliOSS.AccessID, config.Conf.AliOSS.AccessKey,
	)
	bucket, _ := client.Bucket(config.Conf.AliOSS.BucketName)
	return bucket
}

package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type releaseFromExternalContent struct {
	getter                 urlGetter
	rawVersion             string
	artifactConfigurations []buildArtifact
}

func (re releaseFromExternalContent) prepareRelease(baseArchiveUrl, whatsNewUrl string, releaseNotesUrl string, nightly bool) (*release, error) {
	version := re.rawVersion[1:]
	isBeta := strings.Contains(version, "beta")

	builds := []build{}
	for _, ba := range re.artifactConfigurations {
		sha256, err := re.getter.getContents(fmt.Sprintf("%s.sha256", ba.getUrl(baseArchiveUrl, version, isBeta)))
		if err != nil {
			return nil, err
		}
		builds = append(builds, newBuild(baseArchiveUrl, ba, version, isBeta, sha256))
	}

	r := release{
		Version:         version,
		ReleaseDate:     time.Now().UTC(),
		Stable:          !isBeta && !nightly,
		Beta:            isBeta,
		Nightly:         nightly,
		WhatsNewUrl:     whatsNewUrl,
		ReleaseNotesUrl: releaseNotesUrl,
		Builds:          builds,
	}
	return &r, nil
}

type urlGetter interface {
	getContents(url string) (string, error)
}

type getHttpContents struct{}

func (getHttpContents) getContents(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()
	all, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(all), nil
}

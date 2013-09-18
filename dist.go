package dist

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/inconshreveable/go-update"
	"io/ioutil"
	"net/http"
	"runtime"
)

type Dist struct {
	Host    string
	Project string
}

type DistRelease struct {
	Version string
	Url     string
}

func NewDist(host string, project string) (d *Dist) {
	d = new(Dist)
	d.Host = host
	d.Project = project
	return
}

func (d *Dist) Update() (version string, err error) {
	releases, err := d.fetchReleases()
	if len(releases) < 1 {
		return "", errors.New("no releases")
	}
	d.updateFromUrl(releases[0].Url)
	version = releases[0].Version
	return
}

func (d *Dist) UpdateTo(version string) (err error) {
	releases, err := d.fetchReleases()
	for _, release := range releases {
		if release.Version == version {
			d.updateFromUrl(release.Url)
			return
		}
	}
	return errors.New(fmt.Sprintf("no such version: %s", version))
}

func (d *Dist) fetchReleases() (releases []DistRelease, err error) {
	fmt.Printf("host:%s project:%s goos:%s goarch:%s", d.Host, d.Project, runtime.GOOS, runtime.GOARCH)
	url := fmt.Sprintf("%s/projects/%s/releases/%s-%s", d.Host, d.Project, runtime.GOOS, runtime.GOARCH)
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &releases)
	return
}

func (d *Dist) updateFromUrl(url string) (err error) {
	err, _ = update.FromUrl(url)
	return
}

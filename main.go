package main

import (
	"context"
	"fmt"
	"github.com/coreos/go-semver/semver"
	"github.com/google/go-github/github"

	"encoding/csv"
	"os"
	"strings"
	"io"
	"path/filepath"
	"regexp"
)

// LatestVersions returns a sorted slice with the highest version as its first element and the highest version of the smaller minor versions in a descending order
func LatestVersions(releases []*semver.Version, minVersion *semver.Version) []*semver.Version {
	var versionSlice []*semver.Version
	// This is just an example structure of the code, if you implement this interface, the test cases in main_test.go are very easy to run
	// by bruce
	if len(releases) == 0{
		return versionSlice
	}

	// sort O(n * log(n))
	semver.Sort(releases)

	// iterate releases, find target versions
	// O(n)
	index := len(releases) - 1
	target := index
	for index >=0 && releases[index].Compare(*minVersion) >= 0 {
		if index == len(releases) - 1{
			versionSlice = append(versionSlice, releases[index])
		}else if releases[target].Major != releases[index].Major || releases[target].Minor != releases[index].Minor {
			versionSlice = append(versionSlice, releases[index])
			target = index
		}
		index--
	}

	return versionSlice
}

// optimize path
func get_current_path(path string) string{
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println(err.Error())
	}
	return filepath.Join(dir, "backend-challenge", path)
}
func read_file(filename string) [][]string {
	res := [][]string{}

	file, err := os.Open(get_current_path(filename))
	if err != nil {
		fmt.Println("Access file failed", err.Error())
		return res
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var valid_name = regexp.MustCompile(`^.+/.+$`)
	var valid_version = regexp.MustCompile(`^(0|[1-9]+\d*)\.(0|[1-9]+\d*)\.(0|[1-9]+\d*)$`)
	i := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error:", err)
			return res
		}
		//regex filter useless data
		if i == 0 {
			fmt.Printf("data header: %s\n", record)
		} else if len(record) == 2 && valid_name.MatchString(record[0]) && valid_version.MatchString(record[1]) {
			res = append(res, record)
		}else{
			fmt.Printf("data format error: %s\n", record)
		}
		i++
	}
	return res
}

func parse(list [][]string) {
	if len(list) == 0 {
		return
	}

	//by bruce
	//"catch" need to be implemented in case that the whole program crash
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)

		}
	}()

	// Github
	client := github.NewClient(nil)
	ctx := context.Background()
	opt := &github.ListOptions{PerPage: 10}


	// could be optimized with concurrent
	for _, l := range list {
		temp := strings.Split(l[0], "/")
		o, r := temp[0], temp[1]

		//get all pages
		//rate limit may be hit
		var releases []*github.RepositoryRelease
		for {
			repos, resp, err := client.Repositories.ListReleases(ctx, o, r, opt)
			if err != nil {
				panic(err)// is this really a good way? defer,recover
			}
			releases = append(releases, repos...)
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}
		//releases, _, err := client.Repositories.ListReleases(ctx, o, r, opt)
		//if err != nil {
		//	panic(err) // is this really a good way? defer,recover
		//}
		if len(releases) == 0 {
			fmt.Printf("releases of %s can't be accessed by api.github.com\n", l[0])
			continue
		}
		//minVersion := semver.New("1.0.0")
		minVersion := semver.New(l[1])
		allReleases := make([]*semver.Version, len(releases))
		for i, release := range releases {
			versionString := *release.TagName
			if versionString[0] == 'v' {
				versionString = versionString[1:]
			}
			allReleases[i] = semver.New(versionString)
		}
		versionSlice := LatestVersions(allReleases, minVersion)

		fmt.Printf("latest versions of %s after %s : %s\n",l[0], l[1], versionSlice)

	}

}

// Here we implement the basics of communicating with github through the library as well as printing the version
// You will need to implement LatestVersions function as well as make this application support the file format outlined in the README
// Please use the format defined by the fmt.Printf line at the bottom, as we will define a passing coding challenge as one that outputs
// the correct information, including this line
func main() {
	parse(read_file("test.txt"))

}

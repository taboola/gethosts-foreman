/*
 The MIT License (MIT)

Copyright (c) 2013 Tal Sliwowicz

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type HostGetter struct {
	DownloadLocation *string
	CacheLocation    *string
	User             *string
	Password         *string
	CacheFileName    *string
	HostPattern      *string
	HostPrefix       string
	CacheDuration    *time.Duration
}

type Result struct {
	Name string
}
type Response struct {
	Results []Result
}

func (self *HostGetter) downloadHosts() ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", *self.DownloadLocation, nil)
	if err != nil {
		log.Printf("could not create new http request for %s", self.DownloadLocation)
		return nil, err
	}
	req.SetBasicAuth(*self.User, *self.Password)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		log.Printf("could not open http connection to download config from %s", *self.DownloadLocation)
		return nil, err
	}

	data, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Printf("could not download config from %s", self.DownloadLocation)
		return nil, err
	}
	return data, nil
}

func (self *HostGetter) getHosts() (rt string, err error) {
	cacheLoc := filepath.Join(*self.CacheLocation, *self.CacheFileName)
	fi, err := os.Stat(cacheLoc)
	if err != nil {
		log.Printf("no file in cache, downloading %s (%v)", cacheLoc, err)
	} else if !time.Now().After(fi.ModTime().Add(*self.CacheDuration)) {
		log.Printf("found file in cache")
		var data []byte
		data, err = ioutil.ReadFile(cacheLoc)
		if err != nil {
			log.Printf("corrupt file in cache, downloading %s (%v)", cacheLoc, err)
		} else {
			rt = string(data)
			return
		}
	}

	rt, err = self.downloadParseHosts()
	if err == nil {
		log.Printf("downloaded and parsed - trying to save to cache")
		errd := os.MkdirAll(*self.CacheLocation, os.ModePerm)
		if errd != nil {
			log.Printf("could not create cache dir %s (%v)", *self.CacheLocation, errd)
			return
		}
		errf := ioutil.WriteFile(cacheLoc, []byte(rt), os.ModePerm)
		if errf != nil {
			log.Printf("downloaded and parsed ok, could not save to cache %v", errf)
		} else {
			log.Printf("saved hosts in cache")
		}
	} else {
		log.Printf("could not get hosts", err)
	}

	return
}

func (self *HostGetter) downloadParseHosts() (rt string, err error) {
	var data []byte
	var hosts Response

	data, err = self.downloadHosts()
	if err != nil {
		log.Printf("failed to download")
		return
	}
	hosts, err = self.parse(data)
	if err != nil {
		log.Printf("could not parse")
		return
	}
	var buffer bytes.Buffer
	for _, host := range hosts.Results {
		buffer.WriteString(host.Name + "\n")
	}
	rt = buffer.String()
	return
}

func (self *HostGetter) parse(data []byte) (hosts Response, err error) {
	err = json.Unmarshal(data, &hosts)
	if err != nil {
		log.Print("error:", err)
	}
	return
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	log.SetOutput(os.Stderr)

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	getter := HostGetter{
		DownloadLocation: flag.String("url", "https://<forman-host>/hosts", "url to use"),
		CacheLocation:    flag.String("cachedir", filepath.Join(usr.HomeDir, ".gethosts"), "url to use"),
		CacheFileName:    flag.String("cachefile", "hostslist.txt", "url to use"),
		User:             flag.String("user", "", "user name for authentication"),
		Password:         flag.String("password", "", "password"),
		CacheDuration:    flag.Duration("cacheduration", time.Hour, "cache duration before trying to refresh")}

	flag.Parse() // Scan the arguments list

	if len(flag.Args()) > 0 {
		patt := flag.Arg(0)
		i := strings.Index(patt, "@")
		if i != -1 {
			getter.HostPrefix = patt[:i+1]
			patt = patt[i+1:]
		}
		getter.HostPattern = &patt
		log.Printf("requesting pattern %s, for prefix %s", patt, getter.HostPrefix)
	}

	hosts, err := getter.getHosts()
	if err != nil {
		log.Fatal("failed to download hosts")
	}

	if getter.HostPattern != nil {
		hostArr := strings.Split(hosts, "\n")

		for _, line := range hostArr {
			if strings.HasPrefix(line, *getter.HostPattern) {
				fmt.Fprintln(os.Stdout, getter.HostPrefix+line)
			}
		}
	} else {

		fmt.Fprintln(os.Stdout, hosts)
	}
}

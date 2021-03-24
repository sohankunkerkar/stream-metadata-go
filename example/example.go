// Package main contains an example use of this library; it
// prints the current Fedora CoreOS EC2(AWS) x86_64 AMI in the
// us-east-2 region.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/coreos/stream-metadata-go/fedoracoreos"
	"github.com/coreos/stream-metadata-go/stream"
)

const (
	targetArch = "x86_64"
	region     = "us-east-2"
)

func downloadISO(fcosstable stream.Stream) error {
	iso := fcosstable.Architectures[targetArch].Artifacts["metal"].Formats["iso"].Disk
	if iso == nil {
		return fmt.Errorf("%s: missing iso", fcosstable.FormatPrefix(targetArch))
	}
	w, err := os.Create(filepath.Base(iso.Location))
	if err != nil {
		return err
	}
	defer w.Close()
	bufw := bufio.NewWriter(w)

	err = iso.Fetch(w)
	if err != nil {
		return err
	}

	err = bufw.Flush()
	if err != nil {
		return err
	}

	return nil
}

func printAMI(fcosstable stream.Stream) error {
	arch, ok := fcosstable.Architectures[targetArch]
	if !ok {
		return fmt.Errorf("No %s architecture in stream", targetArch)
	}
	awsimages := arch.Images.Aws
	if awsimages == nil {
		return fmt.Errorf("No %s AWS images in stream", targetArch)
	}
	var regionVal stream.AwsRegionImage
	if regionVal, ok = awsimages.Regions[region]; !ok {
		return fmt.Errorf("No %s AWS images in region %s", targetArch, region)
	}
	fmt.Printf("%s\n", regionVal.Image)

	return nil
}

func run() error {
	streamurl := fedoracoreos.GetStreamURL(fedoracoreos.StreamStable)
	resp, err := http.Get(streamurl.String())
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	var fcosstable stream.Stream
	err = json.Unmarshal(body, &fcosstable)
	if err != nil {
		return err
	}
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: example aws-ami|download-iso")
	}
	arg := os.Args[1]
	if arg == "aws-ami" {
		return printAMI(fcosstable)
	} else if arg == "download-iso" {
		return downloadISO(fcosstable)
	} else {
		return fmt.Errorf("invalid operation %s", arg)
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

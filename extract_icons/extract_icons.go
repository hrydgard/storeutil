package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"github.com/hrydgard/storeutil/pbp"
)

/*
type Category struct {

}

type Store struct {
	Type int `json:"type"`
	Version int `json:"version"`
	Categories map[string]Category `json:"categories"`
}
*/

func ExtractData(f *zip.File, subFile int) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	var pbp pbp.PBP
	if err = pbp.Read(rc); err != nil {
		return nil, err
	}
	pbp.Print()
	return pbp.GetSubFile(subFile)
}

func ExtractFile(f *zip.File, subFile int, outFile string) error {
	bytes, err := ExtractData(f, subFile)
	if err != nil {
		return err
	}
	// Write the PNG
	if len(bytes) > 0 {
		return ioutil.WriteFile(outFile, bytes, 0644)
	} else {
		return nil
	}
}

func ExtractImagesFromZip(filename string, outPrefix string) error {
	z, err := zip.OpenReader(filename)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer z.Close()

	found := false
	for _, f := range z.File {
		// fmt.Println("File:", f.Name)
		if strings.HasSuffix(f.Name, "EBOOT.PBP") {
			found = true
			if err := ExtractFile(f, 1, outPrefix+"icon0.png"); err != nil {
				fmt.Println(filename+":", err)
				return err
			}
			if err := ExtractFile(f, 4, outPrefix+"pic1.png"); err != nil {
				fmt.Println(filename+":", err)
				return err
			}
		}
	}
	if !found {
		return errors.New("Did not find EBOOT.PBP in the zip file")
	}
	return nil
}

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		files, err := ioutil.ReadDir("pub/files")
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, info := range files {
			name := info.Name()
			if strings.HasSuffix(name, ".zip") {
				prefix := strings.TrimSuffix(name, ".zip") + "_"
				fmt.Println(prefix)
				ExtractImagesFromZip("pub/files/"+name, prefix)
			} else {
				fmt.Println("Not a zip file:", name)
			}
		}
		return
	}

	filename := args[0]

	err := ExtractImagesFromZip(filename, "")
	if err != nil {
		fmt.Println(err)
	}
}

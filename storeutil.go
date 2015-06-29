package main

import (
	"archive/zip"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func ExtractData(f *zip.File, subFile int) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	var cookie uint32
	var version uint32
	var offsets [8]uint32
	binary.Read(rc, binary.LittleEndian, &cookie)
	if cookie != 0x50425000 {
		return nil, errors.New("bad cookie")
	}

	binary.Read(rc, binary.LittleEndian, &version)
	for i := 0; i < 8; i++ {
		binary.Read(rc, binary.LittleEndian, &offsets[i])
	}

	pngOffset := offsets[subFile]
	pngSize := offsets[subFile+1] - offsets[subFile]

	toSkip := pngOffset - 8*4 - 8
	io.CopyN(ioutil.Discard, rc, int64(toSkip))

	// Alright, now to read the png bytes.
	pngData := make([]byte, pngSize)
	_, err = rc.Read(pngData)
	if err != nil {
		return nil, err
	}

	return pngData, nil
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

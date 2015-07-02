package pbp

import (
	"fmt"
	"os"
	"io"
	"io/ioutil"
	"errors"
	"encoding/binary"
)

const (
	PARAM_SFO = 0
	ICON0_PNG = 1
	ICON1_PMF = 2
	PIC0_PNG = 3
	PIC1_PNG = 4
	SND0_AT3 = 5
	DATA_PSP = 6
	DATA_PSAR = 7
)

var filenames = []string {
	"PARAM.SFO",
	"ICON0.PNG",
	"ICON1.PMF",
	"PIC0.PNG",
	"PIC1.PNG",
	"SND0.AT3",
	"DATA.PSP",
	"DATA.PSAR",
}

type PBP struct {
	isElf bool
	cookie uint32
	version uint32
	offsets [8]uint32
	data [8][]byte
}

func (pbp *PBP) Read(rc io.ReadCloser) error {
	binary.Read(rc, binary.LittleEndian, &pbp.cookie)
	if pbp.cookie == 0x464C457f {
		fmt.Printf("File is an elf, converting to empty PBP")
		bytes, _ := ioutil.ReadAll(rc)
		pbp.data[6] = append([]byte {0x7f, 0x45, 0x4c, 0x46}[:], bytes...)
		pbp.cookie = 0x50425000
		pbp.version = 0x00010000;
		return nil
	}
	if pbp.cookie != 0x50425000 {
		return errors.New("bad cookie")
	}
	binary.Read(rc, binary.LittleEndian, &pbp.version)
	for i := 0; i < 8; i++ {
		binary.Read(rc, binary.LittleEndian, &pbp.offsets[i])
	}

	for i := 0; i < 7; i++ {
		pbp.data[i] = make([]byte, pbp.offsets[i + 1] - pbp.offsets[i])
		if len(pbp.data[i]) > 0 {
			_, err := rc.Read(pbp.data[i])
			if err != nil {
				return err
			}
		}
	}
	var err error
	pbp.data[7], err = ioutil.ReadAll(rc)
	return err
}

func (pbp *PBP) ReadFile(filename string) error {
	rc1, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer rc1.Close()
	return pbp.Read(rc1)
}

func (pbp *PBP) Print() {
	fmt.Printf("Cookie: %08x  Version: %08x\n", pbp.cookie, pbp.version)
	for i := 0; i < 8; i++ {
		fmt.Printf("  %s : %d bytes at offset %d\n", filenames[i], len(pbp.data[i]), pbp.offsets[i])
	}
}

func (pbp *PBP) GetSubFile(i int) ([]byte, error) {
	if i >= 0 && i < 8 {
		return pbp.data[i], nil
	} else {
		return nil, errors.New("No such subfile")
	}
}

func (pbp *PBP) RecalcOffsets() {
	offset := 8 + 8 * 4
	for i := 0; i < 8; i++ {
		pbp.offsets[i] = uint32(offset)
		offset += len(pbp.data[i])
	}
}

func (pbp *PBP) Write(wc io.WriteCloser) {
	pbp.RecalcOffsets()
	binary.Write(wc, binary.LittleEndian, &pbp.cookie)
	binary.Write(wc, binary.LittleEndian, &pbp.version)
	for i := 0; i < 8; i++ {
		binary.Write(wc, binary.LittleEndian, &pbp.offsets[i])
	}
	for i := 0; i < 8; i++ {
		if len(pbp.data[i]) > 0 {
			wc.Write(pbp.data[i])
		}
	}
}

func (pbp *PBP) Merge(other *PBP) error {
	if pbp.cookie != other.cookie {
		return errors.New(fmt.Sprintf("PBP cookies don't match : %08x vs %08x", pbp.cookie, other.cookie))
	}
	if pbp.version != other.version {
		return errors.New(fmt.Sprintf("PBP version don't match : %08x vs %08x", pbp.version, other.version))
	}
	for i := 0; i < 8; i++ {
		if len(other.data[i]) > len(pbp.data[i]) {
			pbp.data[i] = other.data[i]
			fmt.Printf("PBP2 has a better %s - %d vs %d\n", filenames[i], len(other.data[i]), len(pbp.data[i]))
		} else if len(pbp.data[i]) > len(other.data[i]) {
			fmt.Printf("PBP1 has a better %s - %d vs %d\n", filenames[i], len(pbp.data[i]), len(other.data[i]))
		} else {
			fmt.Printf("%s equal - %d vs %d\n", filenames[i], len(pbp.data[i]), len(other.data[i]))
		}
	}
	pbp.RecalcOffsets()
	return nil
}

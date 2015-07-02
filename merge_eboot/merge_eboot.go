package main

import (
	"fmt"
	"os"
	"github.com/hrydgard/storeutil/pbp"
)

func MergePBP(file1, file2 string, output string) error {
	var pbp1, pbp2 pbp.PBP
	pbp1.ReadFile(file1)
	pbp2.ReadFile(file2)
	pbp1.Print()
	pbp2.Print()
	fmt.Println("Merging...")
	if err := pbp1.Merge(&pbp2); err != nil {
		return err
	}

	pbp1.Print()

	fmt.Println("Writing", output)
	wc, err := os.Create(output)
	if err != nil {
		return err
	}
	defer wc.Close()
	pbp1.Write(wc)
	return nil
}

func main() {
	args := os.Args[1:]
	if len(args) != 3 {
		fmt.Println("Wrong number of arguments")
		return
	}

	err := MergePBP(args[0], args[1], args[2])
	if err != nil {
		fmt.Println(err)
		return
	}
}

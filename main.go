package main

import (
	"flag"
	"log"
)

func main() {
	var (
		defFilename string
		outputDir   string
	)

	flag.StringVar(&defFilename, "def", "./tbldef.yaml", "path to table definition yaml file")
	flag.StringVar(&outputDir, "out", "./ddbtbl", "path to output generated table API files")
	flag.Parse()

	err := generateAll(defFilename, outputDir)
	if err != nil {
		log.Fatal(err)
	}
}

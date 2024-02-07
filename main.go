package main

import (
	"flag"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

func main() {
	var (
		defFilename string

		genTableAPI  bool
		tblAPIOutDir string

		genAWSDef    bool
		awsDefOutDir string
	)

	flag.StringVar(&defFilename, "def", "./tbldef.yaml", "path to table definition yaml file")
	flag.BoolVar(&genTableAPI, "api", false, "generate table API")
	flag.StringVar(&tblAPIOutDir, "api_out", "./ddbtbl", "path to output generated table API files")
	flag.BoolVar(&genAWSDef, "aws", false, "generate table definition JSON file for AWS CLI")
	flag.StringVar(&awsDefOutDir, "aws_out", "./awscli_tbldef_json", "path to output generated table definition JSON file for AWS CLI")
	flag.Parse()

	defs, err := loadTableDefinitions(defFilename)
	if err != nil {
		log.Fatal(err)
	}

	if genTableAPI {
		if err := generateTableAPI(defs, tblAPIOutDir); err != nil {
			log.Fatal(err)
		}
	}
	if genAWSDef {
		if err := generateAWSDDBTableDefs(defs, awsDefOutDir); err != nil {
			log.Fatal(err)
		}
	}
}

// read table definition file
func loadTableDefinitions(defFilename string) ([]tableDef, error) {
	defs := make([]tableDef, 0)
	defFile, err := os.Open(defFilename)
	if err != nil {
		return nil, err
	}
	err = yaml.NewDecoder(defFile).Decode(&defs)
	if err != nil {
		return nil, err
	}

	validDefs := make([]tableDef, 0)
	for _, def := range defs {
		if def.isValid() {
			validDefs = append(validDefs, def)
		} else {
			log.Printf("invalid table definition. table name: %s", def.Name)
		}
	}
	return validDefs, nil
}

package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"github.com/gobuffalo/packr/v2"
	"github.com/iancoleman/strcase"
	"gopkg.in/yaml.v2"
)

func parseTemplates() (*template.Template, error) {
	tempFileBox := packr.New("tempFileBox", "./templates")

	temp := template.New("root")
	var err error
	for _, tempName := range tempFileBox.List() {
		tempStr, _ := tempFileBox.FindString(tempName)
		temp, err = temp.New(tempName).Parse(tempStr)
		if err != nil {
			return nil, err
		}
	}

	return temp, nil
}

// remove all contents in outDir if outDir exists
// if not, create ourDir
func cleanUpOutputDir(outDir string) error {
	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		return os.MkdirAll(outDir, 0775)
	}

	files, err := filepath.Glob(filepath.Join(outDir, "*"))
	if err != nil {
		return err
	}
	for _, fpath := range files {
		err := os.RemoveAll(fpath)
		if err != nil {
			return err
		}
	}
	return nil
}

func generateAll(defFilename string, outDir string) error {
	// parse templates
	temp, err := parseTemplates()
	if err != nil {
		return err
	}

	// clean up output directory
	err = cleanUpOutputDir(outDir)
	if err != nil {
		return err
	}

	// write a source file defines logics used by DDB table APIs
	ddbapiSrcPath := path.Join(outDir, "ddbapi.go")
	ddbapiSrcFile, err := os.OpenFile(ddbapiSrcPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	err = temp.ExecuteTemplate(ddbapiSrcFile, "ddbapi_src.gogo", nil)
	if err != nil {
		return err
	}

	// read table definition file
	defs := make([]tableDef, 0)
	defFile, err := os.Open(defFilename)
	if err != nil {
		return err
	}
	err = yaml.NewDecoder(defFile).Decode(&defs)
	if err != nil {
		return err
	}

	// generate table API for each table definition
	for _, tblDef := range defs {
		if !tblDef.isValid() {
			log.Printf("invalid table definition. table name: %s", tblDef.Name)
			continue
		}

		buf, err := generateTableAPI(tblDef.toGenDef(), temp)
		if err != nil {
			log.Println(err)
			continue
		}

		filename := fmt.Sprintf("%s.go", strcase.ToSnake(tblDef.Name))
		outPath := path.Join(outDir, filename)
		err = ioutil.WriteFile(outPath, buf, 0666)
		if err != nil {
			log.Println(err)
			continue
		}
	}
	return nil
}

func generateTableAPI(tbl tableGenDef, temp *template.Template) ([]byte, error) {
	buf := new(bytes.Buffer)

	err := temp.ExecuteTemplate(buf, "tblapi.gogo", tbl)
	if err != nil {
		return nil, err
	}

	return format.Source(buf.Bytes())
}

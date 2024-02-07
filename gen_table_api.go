package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/ettle/strcase"
	"github.com/gobuffalo/packr/v2"
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

// if outDir exists, remove generated files in outDir
// else create ourDir
func cleanUpOutputDir(outDir string) error {
	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		return os.MkdirAll(outDir, 0775)
	}

	files, err := filepath.Glob(pathForGeneratedFile(outDir, "*"))
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

// suffix for path of generated files
const generatedFileSuffix = "gen"

func pathForGeneratedFile(outDir string, name string) string {
	return filepath.Join(outDir, fmt.Sprintf("%s_%s.go", name, generatedFileSuffix))
}

func generateTableAPI(tblDefs []tableDef, outDir string) error {
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
	ddbapiSrcPath := pathForGeneratedFile(outDir, "ddbapi")
	ddbapiSrcFile, err := os.OpenFile(ddbapiSrcPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	err = temp.ExecuteTemplate(ddbapiSrcFile, "ddbapi_src.gogo", nil)
	if err != nil {
		return err
	}

	// generate table API for each table definition
	for _, tblDef := range tblDefs {
		buf, err := generateTableAPISrc(tblDef.toGenDef(), temp)
		if err != nil {
			log.Println(err)
			continue
		}

		apiSrcPath := pathForGeneratedFile(outDir, strcase.ToSnake(tblDef.Name))
		err = ioutil.WriteFile(apiSrcPath, buf, 0666)
		if err != nil {
			log.Println(err)
			continue
		}
	}
	return nil
}

func generateTableAPISrc(tbl tableGenDef, temp *template.Template) ([]byte, error) {
	buf := new(bytes.Buffer)

	err := temp.ExecuteTemplate(buf, "tblapi.gogo", tbl)
	if err != nil {
		return nil, err
	}

	return format.Source(buf.Bytes())
}

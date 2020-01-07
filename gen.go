package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path"
	"text/template"

	"github.com/iancoleman/strcase"
	"gopkg.in/yaml.v2"
)

func generateAll(defFilename string, outDir string) {
	// parse templates
	temp := template.Must(template.ParseGlob("./templates/*.gogo"))

	// create output directory
	err := os.MkdirAll(outDir, 0666)
	if err != nil {
		log.Println(err)
		return
	}

	// write a source file defines logics used by DDB table APIs
	ddbapiSrcPath := path.Join(outDir, "ddbapi.go")
	ddbapiSrcFile, err := os.OpenFile(ddbapiSrcPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		log.Println(err)
		return
	}
	err = temp.ExecuteTemplate(ddbapiSrcFile, "ddbapi_src.gogo", nil)
	if err != nil {
		log.Println(err)
		return
	}

	// read table definition file
	defs := make([]tableDef, 0)
	defFile, err := os.Open(defFilename)
	if err != nil {
		log.Println(err)
		return
	}
	err = yaml.NewDecoder(defFile).Decode(&defs)
	if err != nil {
		log.Println(err)
		return
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
}

func generateTableAPI(tbl tableGenDef, temp *template.Template) ([]byte, error) {
	buf := new(bytes.Buffer)

	err := temp.ExecuteTemplate(buf, "tblapi.gogo", tbl)
	if err != nil {
		return nil, err
	}

	return format.Source(buf.Bytes())
}

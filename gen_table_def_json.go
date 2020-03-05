package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/iancoleman/strcase"
)

type awsDDBTableDef struct {
	TableName              string
	AttributeDefinitions   []awsDDBAttrDef
	KeySchema              awsDDBKeySchema
	GlobalSecondaryIndexes []awsDDBGSIDef `json:",omitempty"`
	BillingMode            string
	StreamSpecification    awsDDBStreamSpec
}

type awsDDBAttrDef struct {
	AttributeName string
	AttributeType string
}

func attrDefToAWSDef(atDef attrDef) awsDDBAttrDef {
	awsTyp := attrType2AWSAttrType[atDef.Type]
	return awsDDBAttrDef{atDef.Name, awsTyp}
}

func extractAWSAttrDefs(tblDef tableDef) []awsDDBAttrDef {
	attrs := make(map[awsDDBAttrDef]struct{})

	attrs[attrDefToAWSDef(tblDef.PrimaryIndex.PK)] = struct{}{}
	if !tblDef.PrimaryIndex.SK.isEmpty() {
		attrs[attrDefToAWSDef(tblDef.PrimaryIndex.SK)] = struct{}{}
	}

	for _, si := range tblDef.SecondaryIndexes {
		attrs[attrDefToAWSDef(si.PK)] = struct{}{}
		if !si.SK.isEmpty() {
			attrs[attrDefToAWSDef(si.SK)] = struct{}{}
		}
	}

	res := make([]awsDDBAttrDef, 0)
	for awsAttr := range attrs {
		res = append(res, awsAttr)
	}
	return res
}

type awsDDBKeyDef struct {
	AttributeName string
	KeyType       string
}

type awsDDBKeySchema []awsDDBKeyDef

func indexDefToKeySchema(idxDef indexDef) awsDDBKeySchema {
	res := make(awsDDBKeySchema, 0)

	res = append(res, awsDDBKeyDef{
		AttributeName: idxDef.PK.Name,
		KeyType:       "HASH",
	})
	if !idxDef.SK.isEmpty() {
		res = append(res, awsDDBKeyDef{
			AttributeName: idxDef.SK.Name,
			KeyType:       "RANGE",
		})
	}
	return res
}

type awsDDBGSIDef struct {
	IndexName  string
	KeySchema  awsDDBKeySchema
	Projection struct {
		ProjectionType string
	}
}

func secondaryIndexDefToAWS(idxName string, idxDef indexDef) awsDDBGSIDef {
	return awsDDBGSIDef{
		IndexName: idxName,
		KeySchema: indexDefToKeySchema(idxDef),
		Projection: struct {
			ProjectionType string
		}{
			ProjectionType: "ALL",
		},
	}
}

type awsDDBStreamSpec struct {
	StreamEnabled  bool
	StreamViewType string `json:",omitempty"`
}

func streamSpec(enabled bool) awsDDBStreamSpec {
	if enabled {
		return awsDDBStreamSpec{
			StreamEnabled:  true,
			StreamViewType: "NEW_AND_OLD_IMAGES",
		}
	} else {
		return awsDDBStreamSpec{
			StreamEnabled: false,
		}
	}
}

var attrType2AWSAttrType = map[string]string{
	"string": "S",
	"int":    "N",
	"int64":  "N",
	"bytes":  "B",
}

func tableDefToAWSDef(tblDef tableDef) awsDDBTableDef {
	awsGSIs := make([]awsDDBGSIDef, 0, len(tblDef.SecondaryIndexes))
	for idxName, idxDef := range tblDef.SecondaryIndexes {
		awsGSIs = append(awsGSIs, secondaryIndexDefToAWS(idxName, idxDef))
	}

	return awsDDBTableDef{
		TableName:              tblDef.Name,
		AttributeDefinitions:   extractAWSAttrDefs(tblDef),
		KeySchema:              indexDefToKeySchema(tblDef.PrimaryIndex),
		GlobalSecondaryIndexes: awsGSIs,
		BillingMode:            "PAY_PER_REQUEST",
		StreamSpecification:    streamSpec(tblDef.StreamEnabled),
	}
}

func cleanUpTableDefsOutputDir(outDir string) error {
	if err := os.RemoveAll(outDir); err != nil {
		return err
	}

	return os.MkdirAll(outDir, 0775)
}

func generateAWSDDBTableDefs(tblDefs []tableDef, outDir string) error {
	if err := cleanUpTableDefsOutputDir(outDir); err != nil {
		return err
	}

	for _, tblDef := range tblDefs {
		func() {
			awsDef := tableDefToAWSDef(tblDef)

			tblDefPath := filepath.Join(outDir, fmt.Sprintf("%s.json", strcase.ToSnake(tblDef.Name)))
			f, err := os.OpenFile(tblDefPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
			if err != nil {
				log.Println(err)
				return
			}
			defer f.Close()

			_ = json.NewEncoder(f).Encode(awsDef)
		}()
	}
	return nil
}

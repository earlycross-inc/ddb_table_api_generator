package main

type tableDef struct {
	Name             string              `yaml:"tableName"`
	PrimaryIndex     indexDef            `yaml:"primaryIndex"`
	SecondaryIndexes map[string]indexDef `yaml:"secondaryIndexes"`
}

func (t tableDef) isValid() bool {
	if t.Name == "" {
		return false
	}
	if !t.PrimaryIndex.isValid() {
		return false
	}
	for _, si := range t.SecondaryIndexes {
		if !si.isValid() {
			return false
		}
	}
	return true
}

type indexDef struct {
	PK attrDef `yaml:"pk"`
	SK attrDef `yaml:"sk"`
}

func (i indexDef) isValid() bool {
	return i.PK.isValid() && i.SK.isValid()
}

func (i indexDef) isSimple() bool {
	return i.PK.isValid() && i.SK.isEmpty()
}

type attrDef struct {
	Name string `yaml:"attrName"`
	Type string `yaml:"attrType"`
}

func (a attrDef) isValid() bool {
	return (a.Name != "" && isValidAttrType(a.Type)) || a.isEmpty()
}

func (a attrDef) isEmpty() bool {
	return a.Name == "" && a.Type == ""
}

var attrType2GoType = map[string]string{
	"string": "string",
	"int":    "int",
	"int64":  "int64",
	"bytes":  "[]byte",
}

func isValidAttrType(t string) bool {
	_, ok := attrType2GoType[t]
	return ok
}

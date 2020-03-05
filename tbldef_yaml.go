package main

// テーブル定義yamlファイルの1つのテーブル定義に対応する構造体
type tableDef struct {
	Name             string              `yaml:"tableName"`
	PrimaryIndex     indexDef            `yaml:"primaryIndex"`
	SecondaryIndexes map[string]indexDef `yaml:"secondaryIndexes"`
	StreamEnabled    bool                `yaml:"streamEnabled"`
}

// テーブル定義を検証し、妥当かどうかを返す
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

// テーブル定義データをコード生成テンプレートに渡すためのデータ構造に変換
func (t tableDef) toGenDef() tableGenDef {
	genScndIdxes := make([]indexGenDef, 0, len(t.SecondaryIndexes))
	for idxName, idx := range t.SecondaryIndexes {
		genScndIdxes = append(genScndIdxes, indexGenDef{
			TblName: caseString(t.Name),
			IdxName: caseString(idxName),
			PK:      idx.PK.toGenDef(),
			SK:      idx.SK.toGenDef(),
		})
	}

	return tableGenDef{
		TblName: caseString(t.Name),
		PrimIdx: indexGenDef{
			TblName: caseString(t.Name),
			PK:      t.PrimaryIndex.PK.toGenDef(),
			SK:      t.PrimaryIndex.SK.toGenDef(),
		},
		ScndIdxes: genScndIdxes,
	}
}

type indexDef struct {
	PK attrDef `yaml:"pk"`
	SK attrDef `yaml:"sk"`
}

func (i indexDef) isValid() bool {
	return i.PK.isValid() && !i.PK.isEmpty() && i.SK.isValid()
}

func (i indexDef) IsSimple() bool {
	return i.PK.isValid() && i.SK.isEmpty()
}

type attrDef struct {
	Name string `yaml:"attrName"`
	Type string `yaml:"attrType"`
}

func (a attrDef) isValid() bool {
	return (a.Name != "" && attrType(a.Type).IsValid()) || a.isEmpty()
}

func (a attrDef) isEmpty() bool {
	return a.Name == "" && a.Type == ""
}

func (a attrDef) toGenDef() attrGenDef {
	return attrGenDef{
		AttrName: caseString(a.Name),
		Type:     attrType(a.Type),
	}
}

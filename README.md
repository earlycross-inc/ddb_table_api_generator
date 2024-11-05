# ddb_table_api_generator

DynamoDBテーブルスキーマ定義から以下を自動生成するツール

- 基本的なテーブルの操作を行うコード
- `aws dynamodb create-table`に渡すJSON形式のテーブル定義

## インストール
```sh
# aws-sdk-go-v2 + dynamo/v2 を利用する場合
go install github.com/earlycross-inc/ddb_table_api_generator/v2@latest

# aws-sdk-go(v1) + dynamo(v1) を利用する場合
go install github.com/earlycross-inc/ddb_table_api_generator@latest
```

## 実行オプション

生成したい対象に対応するオプションを指定する(指定がない場合何も生成しない)。

- `--api`: テーブル操作を行うコード
- `--aws`: `aws` CLIに渡すテーブル定義

`--def <ファイル名>`でテーブル定義ファイルを指定する。デフォルトは `./tbldef.yaml`

## テーブルスキーマ定義の書式

```yaml
# インデックスがパーティションキーのみを含むテーブルの例
- tableName: TUser
  primaryIndex: # プライマリインデックスの定義
    pk: # パーティションキー(PK)の定義
      attrName: uid    # PKの属性名
      attrType: string # 属性のデータ型
  secondaryIndexes: # グローバルセカンダリインデックスの定義
    - name: group-index # インデックス名
      pk:
        attrName: group_id
        attrType: int
  streamEnabled: true # DynamoDB Streamsを有効にするか否か(デフォルト: false)

# インデックスがパーティションキーに加えソートキーを含むテーブルの例
- tableName: TUserHoge
  primaryIndex:
    pk:
      attrName: uid
      attrType: int
    sk: # ソートキー(SK)
      attrName: hoge_id # SKの属性名
      attrType: int     # 属性のデータ型
  secondaryIndexes:
    - name: fuga-index
      pk:
        attrName: fuga
        attrType: string
      sk:
        attrName: fuga_value
        attrType: int

    - name: poyo-index
      pk:
        attrName: poyo
        attrType: string
      sk:
        attrName: poyo_value
        attrType: int64
```
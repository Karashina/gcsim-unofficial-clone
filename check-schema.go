package main

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        "./data/gcsim.db",
	}, &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// インデックス情報を取得
	var indexes []struct {
		Type  string `gorm:"column:type"`
		Name  string `gorm:"column:name"`
		Table string `gorm:"column:tbl_name"`
		SQL   string `gorm:"column:sql"`
	}

	err = db.Raw("SELECT type, name, tbl_name, sql FROM sqlite_master WHERE type='index' AND tbl_name='users'").Scan(&indexes).Error
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Indexes on 'users' table ===")
	for _, idx := range indexes {
		fmt.Printf("Name: %s\n", idx.Name)
		fmt.Printf("SQL: %s\n\n", idx.SQL)
	}

	// スキーマを取得
	var schema string
	err = db.Raw("SELECT sql FROM sqlite_master WHERE type='table' AND name='users'").Scan(&schema).Error
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Table schema ===")
	fmt.Println(schema)
}

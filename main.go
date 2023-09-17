package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	DbConn *sql.DB

	mysqlTextTypes   = []string{"text", "varchar", "json", "char"}
	mysqlIntTypes    = []string{"tinyint", "smallint", "mediumint", "integer", "int", "bigint", "bool", "boolean"}
	mysqlFloatTypes  = []string{"float", "double", "decimal", "dec"}
	mysqlBinaryTypes = []string{"tinyblob", "blob", "mediumblob", "longblob"}
	dbName           = flag.String("db", "", "db name")
	tableName        = flag.String("t", "", "table name")
	modelPackage     = flag.String("mp", "main", "model package name")
)

type Column struct {
	Field   string
	Type    string
	Null    string
	Key     string
	Default string
	Extra   string
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	tablesString := os.Getenv("TABLES")
	flag.Parse()

	tables := []string{}

	prepareDB()

	defer DbConn.Close()

	if tablesString != "" {
		tables = strings.Split(tablesString, ",")
	} else if *tableName != "" {
		tables = append(tables, *tableName)
	} else {
		log.Fatal("use -t flag to select table name")
	}

	for _, table := range tables {

		fmt.Println("Table from list: " + table)

		rows, err := DbConn.Query("SHOW COLUMNS FROM " + table)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		columns := []Column{}

		for rows.Next() {
			col := Column{}
			rows.Scan(&col.Field, &col.Type, &col.Null, &col.Key, &col.Default, &col.Extra)
			columns = append(columns, col)
		}

		file, fileErr := os.Create(table + ".go")
		defer file.Close()
		if fileErr != nil {
			log.Fatal(fileErr)
		}

		modelName := titleToUpper(table)

		fileContent := fmt.Sprintf("package %s\n\ntype %s struct {\n", *modelPackage, modelName)

		for _, c := range columns {
			fieldType := getFieldType(c.Type)
			line := fmt.Sprintf("\t%s %s\n", titleToUpper(c.Field), fieldType)
			fileContent = fileContent + line
		}

		fileContent = fileContent + fmt.Sprintf("}\n")
		file.WriteString(fileContent)
	}
}

func titleToUpper(str string) string {
	caser := cases.Title(language.AmericanEnglish)
	return caser.String(str)
}

func prepareDB() {

	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbCreds := fmt.Sprintf("%s:%s@/%s", dbUser, dbPass, dbName)

	conn, err := sql.Open("mysql", dbCreds)
	if err != nil {
		panic(err)
	}
	DbConn = conn

	err = DbConn.Ping()
	if err != nil {
		log.Fatal(err)
	}
}

func getFieldType(mysqlType string) string {
	for _, t := range mysqlTextTypes {
		if strings.HasPrefix(mysqlType, t) {
			return "string"
		}
	}

	for _, t := range mysqlIntTypes {
		if strings.HasPrefix(mysqlType, t) {
			return "int"
		}
	}

	for _, t := range mysqlFloatTypes {
		if strings.HasPrefix(mysqlType, t) {
			return "float64"
		}
	}

	for _, t := range mysqlBinaryTypes {
		if strings.HasSuffix(mysqlType, t) {
			return "[]byte"
		}
	}

	return ""
}

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"go.mercari.io/datastore"
	"go.mercari.io/datastore/clouddatastore"
)

const (
	sqlHeader = `-- DDL generated by ds2dd. DO NOT EDIT!

`
	nestedKeyDelimiter = "__"
)

type prop struct {
	Repr []string `datastore:"property_representation"`
}

// represents map[table name][column name] = [<type1>, <type2>, ...]
type propertyTypes map[string]map[string][]string

// getProperties returns table/column/types from actual Datastore properties.
func getPropertyTypes(ctx context.Context, client datastore.Client) propertyTypes {
	query := client.NewQuery("__property__")

	var props []prop
	keys, err := client.GetAll(ctx, query, &props)
	if err != nil {
		log.Fatalf("client.GetAll: %v", err)
	}

	propNum := len(props)
	keyNum := len(keys)
	if propNum != keyNum {
		log.Fatalf("Invalid result: props %d values, keys %d values.", propNum, keyNum)
	}

	properties := make(propertyTypes)
	for i := 0; i < propNum; i++ {
		k := keys[i]
		repr := props[i].Repr

		tableName := k.ParentKey().Name()
		colName := strings.Replace(k.Name(), ".", nestedKeyDelimiter, -1)

		if properties[tableName] == nil {
			properties[tableName] = make(map[string][]string)
		}

		properties[tableName][colName] = repr
	}

	return properties
}

func format(t propertyTypes) string {
	var sqlStr string

	tableNames := make([]string, 0, len(t))
	for n, _ := range t {
		tableNames = append(tableNames, n)
	}
	sort.Strings(tableNames)

	for _, tableName := range tableNames {
		cols, tableOk := t[tableName]
		if !tableOk {
			fmt.Fprintf(os.Stderr, "Unknown table entry. key name : %s\n", tableName)
			continue
		}

		colNames := make([]string, 0, len(cols))
		for n, _ := range cols {
			colNames = append(colNames, n)
		}
		sort.Strings(colNames)

		var columns []string
		for _, colName := range colNames {
			colTypes, columnOk := cols[colName]
			if !columnOk {
				fmt.Fprintf(os.Stderr, "Unknown column entry. key name : %s\n", colName)
				continue
			}
			if len(colTypes) > 1 {
				fmt.Fprintf(os.Stderr, "Type of %s is ambiguous. A first value is selected. : acceptable types are %v\n", colName, colTypes)
				continue
			}

			t := propRepr2mysqlType(colTypes[0])
			columns = append(columns, fmt.Sprintf("%s %s", colName, t))
		}

		sqlStr += fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s ( \n  %s \n);\n", tableName, strings.Join(columns, ",\n  "))
	}

	return sqlStr
}

// Convert property_representation to MySQL type.
// See detail about the repr on https://cloud.google.com/datastore/docs/concepts/metadataqueries#property_queries_property_representations
func propRepr2mysqlType(propRepr string) string {
	switch propRepr {
	case "INT64":
		return "BIGINT"
	case "DOUBLE":
		return "DOUBLE"
	case "BOOLEAN":
		return "BOOLEAN"
	case "STRING":
		return "VARCHAR(255)"
	default:
		fmt.Fprintf(os.Stderr, "Non convertable type : %s\n", propRepr)
		os.Exit(-1)
	}
	return ""
}

func main() {
	ctx := context.Background()

	// Pass config via env.
	// If you use Datastore emulator, get it by 'gcloud beta emulators datastore env-init'
	client, err := clouddatastore.FromContext(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	properties := getPropertyTypes(ctx, client)
	sqlStr := sqlHeader + format(properties)

	fmt.Println(sqlStr)
}

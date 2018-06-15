package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"go.mercari.io/datastore"
	"go.mercari.io/datastore/clouddatastore"
)

type prop struct {
	Repr []string `datastore:"property_representation"`
}

type propertyTypes map[string]map[string][]string

// getProperties returns map[table name][column name] = [<type1>, <type2>] from Datastore properties.
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

		name := k.Name()
		startOfColumn := strings.Index(name, ".")
		if startOfColumn == -1 {
			continue // skip it
		}

		tableName := name[:startOfColumn]
		colName := strings.Replace(name[startOfColumn+1:], ".", "__", -1)

		if properties[tableName] == nil {
			properties[tableName] = make(map[string][]string)
		}

		properties[tableName][colName] = repr
	}

	return properties
}

func format(t propertyTypes) string {
	var sqlStr string

	for tableName, cols := range t {
		var columns []string
		for colName, colTypes := range cols {
			// FIXME check an available type
			t := colTypes[0]

			columns = append(columns, fmt.Sprintf("%s %s", colName, t))
		}

		sqlStr += fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s ( %s );\n", tableName, strings.Join(columns, ","))
	}

	return sqlStr
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
	sqlStr := format(properties)
	fmt.Println(sqlStr)
}

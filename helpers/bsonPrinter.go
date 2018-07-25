package helpers

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"

	"gopkg.in/mgo.v2/bson"
)

// CloneBsonMap clone bson.M
func CloneBsonMap(m bson.M) (bson.M, error) {
	gob.Register(bson.M{})
	gob.Register(bson.RegEx{})

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	err := enc.Encode(m)
	if err != nil {
		return nil, err
	}

	var copy bson.M
	err = dec.Decode(&copy)
	if err != nil {
		return nil, err
	}

	return copy, nil
}

// CloneBsonMapSlice clone bson.M slice
func CloneBsonMapSlice(ms []bson.M) ([]bson.M, error) {
	result := []bson.M{}
	for _, m := range ms {
		toMap, err := CloneBsonMap(m)
		if err != nil {
			return nil, err
		}
		result = append(result, toMap)
	}

	return result, nil
}

// PrintBsonSlice is print bson slice
func PrintBsonSlice(slice []bson.M, msg string) {
	fmt.Println(fmt.Sprintf("--- %s ---", msg))
	for i, b := range slice {
		fmt.Println("{")
		printBson(b, 0)
		str := "}"
		if i < len(slice)-1 {
			str += ","
		}
		fmt.Println(str)
	}
}

// PrintBson is print bson
func PrintBson(b bson.M, msg string) {
	fmt.Println(fmt.Sprintf("--- %s ---", msg))
	fmt.Println("{")
	printBson(b, 0)
	fmt.Println("}")
}

func printBson(b bson.M, indent int) {
	// printBsonIndent(indent)
	if indent > 0 {
		fmt.Println("{")
	}
	indent++
	for k, v := range b {
		printBsonIndent(indent)
		fmt.Print(fmt.Sprintf("\"%s\":", k))
		printBsonValue(v, indent)
	}
	printBsonIndent(indent - 1)
	if indent-1 > 0 {
		fmt.Println("},")
	}
}

func printBsonD(slice bson.D, indent int) {
	if indent > 0 {
		fmt.Println("{")
	}
	indent++
	for _, b := range slice {
		printBsonIndent(indent)
		printBsonValue(b, indent)
	}
	printBsonIndent(indent - 1)
	if indent-1 > 0 {
		fmt.Println("},")
	}
}

func printDocElem(doc bson.DocElem, indent int) {
	// if indent > 0 {
	// 	fmt.Println("{")
	// }
	// indent++

	// printBsonIndent(indent)
	fmt.Print(fmt.Sprintf("\"%s\":", doc.Name))
	printBsonValue(doc.Value, indent)
	// printBsonIndent(indent - 1)
	// if indent-1 > 0 {
	// 	fmt.Println("},")
	// }
}

func printBsonValue(v interface{}, indent int) {
	switch v.(type) {
	case bool:
		fmt.Println(fmt.Sprintf("%t,", v.(bool)))
	case int, int8, int16, int32, int64, float32, float64:
		fmt.Println(fmt.Sprintf("%d,", v))
	case string:
		fmt.Println(fmt.Sprintf("\"%s\",", v))
	case []string:
		fmt.Print("[")
		for i, s := range v.([]string) {
			format := "\"%s\""
			if i < len(v.([]string))-1 {
				format += ","
			}
			fmt.Print(fmt.Sprintf(format, s))
		}
		fmt.Println("]")
	case bson.M:
		printBson(v.(bson.M), indent)
	case bson.D:
		printBsonD(v.(bson.D), indent)
	case bson.DocElem:
		printDocElem(v.(bson.DocElem), indent)
	default:
		fmt.Println(fmt.Sprintf("---type:%s---", reflect.TypeOf(v).String()))
	}
}

func printBsonIndent(indent int) {
	const bsonPrintIndent = "    "

	for i := 0; i < indent; i++ {
		fmt.Print(bsonPrintIndent)
	}
}

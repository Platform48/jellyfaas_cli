package entities

import (
	"encoding/json"
	"reflect"
	"strings"
)

func GenerateJsonSchemaFromJsonString(jsonString string, flat bool) (*string, error) {
	var data interface{}

	dec := json.NewDecoder(strings.NewReader(jsonString))
	dec.UseNumber() // Use json.Number instead of float64 for numbers

	err := dec.Decode(&data)
	if err != nil {
		return nil, err
	}

	// Generate the schema for the JSON data
	schemaProperties, required := generateSchema(data)

	// Add the additional properties to the root schema
	rootSchema := map[string]interface{}{
		"$schema":    "https://json-schema.org/draft/2020-12/schema",
		"title":      "Generated schema from jellyfaas",
		"type":       "object",
		"properties": schemaProperties["properties"],
		"required":   required,
	}

	if flat {
		schemaJSON, err := json.Marshal(rootSchema)
		if err != nil {
			return nil, err
		}
		ret := string(schemaJSON)
		return &ret, nil
	}

	schemaJSON, err := json.MarshalIndent(rootSchema, "", "  ")
	if err != nil {
		return nil, err
	}

	ret := string(schemaJSON)
	return &ret, nil
}

func generateSchema(data interface{}) (map[string]interface{}, []string) {
	schema := map[string]interface{}{
		"type": getJSONType(data),
	}
	var required []string

	switch v := data.(type) {
	case map[string]interface{}:
		properties := make(map[string]interface{})
		for key, value := range v {
			propertySchema, _ := generateSchema(value)
			properties[key] = propertySchema
			required = append(required, key)
		}
		schema["properties"] = properties
		if len(required) > 0 {
			schema["required"] = required
		}
	case []interface{}:
		if len(v) > 0 {
			itemSchema, _ := generateSchema(v[0])
			schema["items"] = itemSchema
		}
	}

	return schema, required
}

func getJSONType(v interface{}) string {

	switch v := v.(type) {
	case json.Number:
		// Try parsing as integer
		if _, err := v.Int64(); err == nil {
			return "integer"
		} else if _, err := v.Float64(); err == nil {
			return "number"
		}
	}

	switch reflect.TypeOf(v).Kind() {
	case reflect.Map:
		return "object"
	case reflect.Slice:
		return "array"
	case reflect.Int:
		return "integer"
	case reflect.Int32:
		return "integer"
	case reflect.Int64:
		return "integer"
	case reflect.Float32:
		return "number"
	case reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.String:
		return "string"
	default:
		return "null"
	}
}

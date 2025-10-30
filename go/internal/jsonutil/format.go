package jsonutil

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// PrettyJSON formats a Go struct as indented JSON with human-readable timestamps
func PrettyJSON(input interface{}) string {
	// First convert to map to manipulate the data
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		return fmt.Sprintf("Error marshaling JSON: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return fmt.Sprintf("Error unmarshaling JSON: %v", err)
	}

	// Process the map to convert timestamps
	processMap(data)

	// Convert back to pretty JSON
	prettyJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error creating pretty JSON: %v", err)
	}

	return string(prettyJSON)
}

// PrettyJSONPB formats a Protocol Buffer message as indented JSON with human-readable timestamps
func PrettyJSONPB(input proto.Message) string {
	// Use standard Google protojson marshaler
	marshaler := protojson.MarshalOptions{
		Indent:          "  ",
		UseProtoNames:   true,
		EmitUnpopulated: false,
	}

	jsonBytes, err := marshaler.Marshal(input)
	if err != nil {
		return fmt.Sprintf("Error marshaling Protocol Buffer: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return fmt.Sprintf("Error unmarshaling JSON: %v", err)
	}

	// Process the map to convert timestamps
	processMap(data)

	// Convert back to pretty JSON
	prettyJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error creating pretty JSON: %v", err)
	}

	return string(prettyJSON)
}

// processMap recursively processes a map to convert timestamp objects
func processMap(data map[string]interface{}) {
	for key, value := range data {
		switch v := value.(type) {
		case map[string]interface{}:
			// Check if this map might be a timestamp (has seconds and nanos fields)
			if seconds, hasSeconds := v["seconds"]; hasSeconds {
				if nanos, hasNanos := v["nanos"]; hasNanos {
					// Try to convert to a timestamp
					if sec, ok := seconds.(string); ok {
						if nano, ok := nanos.(float64); ok {
							if secInt, err := strconv.ParseInt(sec, 10, 64); err == nil {
								// Create a timestamp in RFC3339 format
								t := time.Unix(secInt, int64(nano)).Format(time.RFC3339)
								data[key] = t
								continue
							}
						}
					}
				}
			}
			// Recursively process nested maps
			processMap(v)
		case []interface{}:
			// Recursively process arrays
			for i, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					processMap(itemMap)
					v[i] = itemMap
				}
			}
		}
	}
}

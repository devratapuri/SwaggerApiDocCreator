package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
)

// Define the basic Swagger structure
type SwaggerTemplate struct {
	OpenAPI string                          `yaml:"openapi"`
	Info    map[string]interface{}          `yaml:"info"`
	Paths   map[string]map[string]Operation `yaml:"paths"`
}

type Operation struct {
	Summary     string               `yaml:"summary"`
	Responses   map[string]Response  `yaml:"responses"`
	Description string               `yaml:"description"`
}

type Response struct {
	Description string               `yaml:"description"`
	Content     map[string]MediaType `yaml:"content"`
}

type MediaType struct {
	Schema Schema `yaml:"schema"`
}

type Schema struct {
	Type       string            `yaml:"type"`
	Properties map[string]Schema `yaml:"properties,omitempty"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Ask user for the desired action
		fmt.Print("\nEnter action (view/create/update/exit): ")
		action, _ := reader.ReadString('\n')
		action = strings.ToLower(strings.TrimSpace(action))

		// Exit the program if the user types 'exit'
		if action == "exit" {
			fmt.Println("Exiting program.")
			break
		}

		// Handle the desired action
		switch action {
		case "view":
			fmt.Print("Enter the path to the Swagger YAML file: ")
			filePath, _ := reader.ReadString('\n')
			filePath = strings.TrimSpace(filePath)
			err := viewSwagger(filePath)
			if err != nil {
				fmt.Println("Error viewing Swagger file:", err)
			}
		case "create":
			fmt.Print("Enter the path to create a new Swagger YAML file: ")
			filePath, _ := reader.ReadString('\n')
			filePath = strings.TrimSpace(filePath)
			err := createSwagger(filePath)
			if err != nil {
				fmt.Println("Error creating Swagger file:", err)
			}
		case "update":
			fmt.Print("Enter the path to the Swagger YAML file: ")
			filePath, _ := reader.ReadString('\n')
			filePath = strings.TrimSpace(filePath)
			err := updateSwagger(filePath, reader)
			if err != nil {
				fmt.Println("Error updating Swagger file:", err)
			}
		default:
			fmt.Println("Invalid action. Please enter 'view', 'create', 'update', or 'exit'.")
		}
	}
}

// View an existing Swagger YAML file
func viewSwagger(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	fmt.Println("Swagger File Contents:")
	fmt.Println(string(data))
	return nil
}

// Create a new Swagger YAML file with a basic structure
func createSwagger(filePath string) error {
	swagger := SwaggerTemplate{
		OpenAPI: "3.0.3",
		Info: map[string]interface{}{
			"title":       "New API",
			"description": "This is a newly created Swagger API",
			"version":     "1.0.0",
		},
		Paths: make(map[string]map[string]Operation),
	}

	return writeSwaggerFile(filePath, &swagger)
}

// Update an existing Swagger YAML file
func updateSwagger(filePath string, reader *bufio.Reader) error {
	// Read existing Swagger YAML file
	swagger, err := readSwaggerFile(filePath)
	if err != nil {
		return err
	}

	// Adding or updating an existing path based on user input
	fmt.Print("Enter the path to add/update (e.g., /pets): ")
	path, _ := reader.ReadString('\n')
	path = strings.TrimSpace(path)

	fmt.Print("Enter HTTP method (get/post/put/delete): ")
	method, _ := reader.ReadString('\n')
	method = strings.ToLower(strings.TrimSpace(method))

	// Prompt user to provide JSON response as a string or a file path
	fmt.Print("Enter JSON response directly or type 'file' to provide a file path: ")
	inputType, _ := reader.ReadString('\n')
	inputType = strings.TrimSpace(strings.ToLower(inputType))

	var jsonData map[string]interface{}

	if inputType == "file" {
		// User wants to provide a file path
		fmt.Print("Enter the JSON file path: ")
		jsonFilePath, _ := reader.ReadString('\n')
		jsonFilePath = strings.TrimSpace(jsonFilePath)

		fileData, err := ioutil.ReadFile(jsonFilePath)
		if err != nil {
			return err
		}

		err = json.Unmarshal(fileData, &jsonData)
		if err != nil {
			return err
		}
	} else {
		// User provides JSON directly
		err = json.Unmarshal([]byte(inputType), &jsonData)
		if err != nil {
			return err
		}
	}

	// Generate the schema from JSON
	schema := generateSchema(jsonData)

	// Check if the path and method already exist
	if swagger.Paths == nil {
		swagger.Paths = make(map[string]map[string]Operation)
	}
	if swagger.Paths[path] == nil {
		swagger.Paths[path] = make(map[string]Operation)
	}

	// Update the existing operation or create a new one
	if existingOperation, ok := swagger.Paths[path][method]; ok {
		// If operation already exists, update the response
		fmt.Println("Updating the existing operation response...")
		existingOperation.Responses["200"] = Response{
			Description: "Successful response",
			Content: map[string]MediaType{
				"application/json": {
					Schema: schema,
				},
			},
		}
		swagger.Paths[path][method] = existingOperation
	} else {
		// Create a new operation if it does not exist
		fmt.Println("Creating a new operation...")
		newOperation := Operation{
			Summary:     "Sample operation for " + path,
			Description: "This is a sample description for the new operation.",
			Responses: map[string]Response{
				"200": {
					Description: "Successful response",
					Content: map[string]MediaType{
						"application/json": {
							Schema: schema,
						},
					},
				},
			},
		}
		swagger.Paths[path][method] = newOperation
	}

	// Write the updated Swagger YAML back to the file
	return writeSwaggerFile(filePath, swagger)
}

// Generate a Swagger schema from a JSON object
func generateSchema(data map[string]interface{}) Schema {
	schema := Schema{Type: "object", Properties: make(map[string]Schema)}

	for key, value := range data {
		fieldType := reflect.TypeOf(value).Kind()
		propSchema := Schema{Type: getSwaggerType(fieldType)}
		if fieldType == reflect.Map {
			propSchema.Type = "object"
		} else if fieldType == reflect.Slice {
			propSchema.Type = "array"
		}
		schema.Properties[key] = propSchema
	}

	return schema
}

// Get Swagger-compatible type from Go's reflect kind
func getSwaggerType(kind reflect.Kind) string {
	switch kind {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int32, reflect.Int64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice:
		return "array"
	case reflect.Map:
		return "object"
	default:
		return "string"
	}
}

// Read an existing Swagger YAML file
func readSwaggerFile(filename string) (*SwaggerTemplate, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var swagger SwaggerTemplate
	err = yaml.Unmarshal(data, &swagger)
	if err != nil {
		return nil, err
	}

	return &swagger, nil
}

// Write Swagger YAML file
func writeSwaggerFile(filename string, swagger *SwaggerTemplate) error {
	data, err := yaml.Marshal(swagger)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}

	fmt.Println("Swagger file updated successfully.")
	return nil
}

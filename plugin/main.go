package main

import (
	"fmt"
	"strings"
	"github.com/spf13/viper"
)

// processMap recursively processes YAML data
func processMap(i interface{}, level int) {
    indent := "  " // Indentation for pretty printing
    switch v := i.(type) {
    case map[string]any:
        for key, value := range v {
            fmt.Printf("%s%v:\n", strings.Repeat(indent, level), key)
            processMap(value, level+1)
        }
    case map[interface{}]interface{}:
        for key, value := range v {
            fmt.Printf("%s%v:\n", strings.Repeat(indent, level), key)
            processMap(value, level+1)
        }
    case []interface{}:
        for _, value := range v {
            processMap(value, level+1)
        }
    default:
        // Handle base types (int, string, bool, etc.)
	fmt.Printf("%s%v\n", strings.Repeat(indent, level), v)
    }
}

func main() {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml") // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("config/")   // path to look for the config file in
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	fmt.Println("CURRENT MIDDLEWARES:")
	for name,middleware := range viper.Get("http.middlewares").(map[string]any) {
		fmt.Printf("%s:\n", name)
		for mtype,_ := range middleware.(map[string]any) {
			fmt.Printf("  %s:\n", mtype)
			fmt.Printf("    %v:\n", viper.Get(fmt.Sprintf("http.middlewares.%s.%s", name,mtype)))
		}
	}

	fmt.Println("\nCustom walk the middlewares.")
	processMap(viper.Get("http.middlewares"), 0)

	fmt.Println("Current routers:")
	processMap(viper.Get("http.routers"),0)

	fmt.Println("Getting specific router info")
	fmt.Println("console:")
	fmt.Printf("EntryPoints:")
	processMap(viper.Get("http.routers.console.entryPoints"),0)
	fmt.Printf("Service:")
	processMap(viper.Get("http.routers.console.service"),0)
	fmt.Printf("Rule:")
	processMap(viper.Get("http.routers.console.rule"),0)
	fmt.Printf("Middelware Name:")
	processMap(viper.Get("http.routers.console.middlewares"),0)
	fmt.Println("Middleware Info: secured")
	processMap(viper.Get("http.middlewares.secured"),0)
}

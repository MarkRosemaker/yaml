To install the package, use the following command:

```sh
go get github.com/MarkRosemaker/yaml
```


Here is a basic example of how to use the package:

```go
package main

import (
	"fmt"
	"github.com/MarkRosemaker/yaml"
)

type Example struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	ex := Example{Name: "John Doe", Age: 30}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(ex)
	if (err != nil) {
		fmt.Println("Error marshalling to YAML:", err)
		return
	}
	fmt.Println("YAML Data:", string(yamlData))

	// Unmarshal from YAML
	var ex2 Example
	err = yaml.Unmarshal(yamlData, &ex2)
	if (err != nil) {
		fmt.Println("Error unmarshalling from YAML:", err)
		return
	}
	fmt.Println("Unmarshalled Struct:", ex2)
}
```

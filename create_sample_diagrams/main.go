package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var samples = map[string]string{
	"sample_flow.mmd":     "flowchart TD\n  start([Start]) --> validate{Valid?}\n  validate -- yes --> done([Done])\n  validate -- no --> fix[Fix input]\n  fix --> validate\n",
	"sample_graph.dot":    "digraph G {\n  rankdir=LR;\n  node [shape=box, style=rounded];\n  app -> api;\n  api -> db;\n}\n",
	"sample_sequence.puml": "@startuml\nactor User\nparticipant API\nUser -> API: Request\nAPI --> User: Response\n@enduml\n",
	"sample.svg":          `<svg xmlns="http://www.w3.org/2000/svg" width="200" height="80" role="img"><title>Sample</title><rect x="10" y="10" width="180" height="60" fill="white" stroke="black"/><text x="100" y="45" text-anchor="middle">Sample</text></svg>` + "\n",
}

func main() {
	outDir := filepath.Join(".", "samples")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
		os.Exit(1)
	}

	for name, content := range samples {
		path := filepath.Join(outDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", name, err)
			os.Exit(1)
		}
		fmt.Println(path)
	}
}

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func generateFullLib(cloud, appids, refs string) string {
	lines := []string{
		"<root>",
		" <libs>",
		fmt.Sprintf("  <lib>%s</lib>", refs),
		" </libs>",
		fmt.Sprintf(" <appIds>%s</appIds>", appids),
		fmt.Sprintf(" <cloud>%s</cloud>", cloud),
		"</root>",
	}
	return strings.Join(lines, "\r\n")
}

func generateGroupLib(appids, refs string) string {
	lines := []string{
		"<root>",
		" <libs>",
		fmt.Sprintf("  <lib>%s</lib>", refs),
		" </libs>",
		fmt.Sprintf(" <appIds>%s</appIds>", appids),
		"</root>",
	}
	return strings.Join(lines, "\r\n")
}

func generatePatchLib(refs string) string {
	lines := []string{
		"<root>",
		"        <libs>",
		fmt.Sprintf("                <lib>%s</lib>", refs),
		"        </libs>",
		"</root>",
	}
	return strings.Join(lines, "\n") + "\n"
}

func main() {
	typ := flag.String("type", "", "lib type: full, group, patch")
	cloud := flag.String("cloud", "", "cloud domain (full only)")
	appids := flag.String("appids", "", "comma-separated appIds")
	refs := flag.String("refs", "", "comma-separated references")
	output := flag.String("output", "", "output file path")
	flag.Parse()

	if *typ == "" {
		fmt.Fprintln(os.Stderr, "Error: --type is required")
		os.Exit(1)
	}

	if *refs == "" {
		fmt.Fprintln(os.Stderr, "Error: --refs is required")
		os.Exit(1)
	}

	if *output == "" {
		fmt.Fprintln(os.Stderr, "Error: --output is required")
		os.Exit(1)
	}

	if *typ == "full" && *cloud == "" {
		fmt.Fprintln(os.Stderr, "Error: full type requires --cloud")
		os.Exit(1)
	}

	if (*typ == "full" || *typ == "group") && *appids == "" {
		fmt.Fprintf(os.Stderr, "Error: %s type requires --appids\n", *typ)
		os.Exit(1)
	}

	var content string
	switch *typ {
	case "full":
		content = generateFullLib(*cloud, *appids, *refs)
	case "group":
		content = generateGroupLib(*appids, *refs)
	case "patch":
		content = generatePatchLib(*refs)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown type: %s\n", *typ)
		os.Exit(1)
	}

	outputDir := filepath.Dir(*output)
	if outputDir != "" && outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
			os.Exit(1)
		}
	}

	if err := os.WriteFile(*output, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 已生成: %s\n", *output)
	fmt.Printf("   类型: %s\n", *typ)
	if *typ == "full" {
		fmt.Printf("   <cloud>: %s\n", *cloud)
	}
	if *appids != "" {
		fmt.Printf("   <appIds>: %s\n", *appids)
	}
	fmt.Printf("   <lib> 引用: %s\n", *refs)

	if info, err := os.Stat(*output); err == nil {
		fmt.Printf("   文件大小: %d bytes\n", info.Size())
	}
}

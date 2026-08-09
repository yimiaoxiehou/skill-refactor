package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var mermaidExts = map[string]bool{".mmd": true, ".mermaid": true}
var graphvizExts = map[string]bool{".dot": true, ".gv": true}
var plantumlExts = map[string]bool{".puml": true, ".plantuml": true}
var svgExts = map[string]bool{".svg": true}

func run(cmd []string, cwd string) error {
	fmt.Println("$ " + strings.Join(cmd, " "))
	c := exec.Command(cmd[0], cmd[1:]...)
	if cwd != "" {
		c.Dir = cwd
	}
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func inferKind(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if mermaidExts[ext] {
		return "mermaid"
	}
	if graphvizExts[ext] {
		return "graphviz"
	}
	if plantumlExts[ext] {
		return "plantuml"
	}
	if svgExts[ext] {
		return "svg"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "mermaid"
	}
	text := strings.TrimSpace(string(data))
	if strings.HasPrefix(text, "@startuml") {
		return "plantuml"
	}
	if strings.HasPrefix(text, "digraph") || strings.HasPrefix(text, "graph") {
		return "graphviz"
	}
	if strings.HasPrefix(text, "<svg") {
		return "svg"
	}
	return "mermaid"
}

func renderMermaid(src, out, format string) error {
	mmdc, err := exec.LookPath("mmdc")
	if err != nil {
		return fmt.Errorf("Mermaid rendering requires `mmdc`. Install with `npm install -g @mermaid-js/mermaid-cli`")
	}
	return run([]string{mmdc, "-i", src, "-o", out}, "")
}

func renderGraphviz(src, out, format string) error {
	dot, err := exec.LookPath("dot")
	if err != nil {
		return fmt.Errorf("Graphviz rendering requires `dot`. Install Graphviz")
	}
	return run([]string{dot, "-T" + format, src, "-o", out}, "")
}

func renderPlantUML(src, out, format string) error {
	plantuml, _ := exec.LookPath("plantuml")
	if plantuml != "" {
		tmpDir, err := os.MkdirTemp("", "plantuml-*")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpDir)
		srcData, _ := os.ReadFile(src)
		tmpSrc := filepath.Join(tmpDir, filepath.Base(src))
		os.WriteFile(tmpSrc, srcData, 0644)
		return run([]string{plantuml, "-t" + format, tmpSrc}, tmpDir)
	}
	java, _ := exec.LookPath("java")
	jar := os.Getenv("PLANTUML_JAR")
	if java != "" && jar != "" {
		if _, err := os.Stat(jar); err == nil {
			tmpDir, _ := os.MkdirTemp("", "plantuml-*")
			defer os.RemoveAll(tmpDir)
			srcData, _ := os.ReadFile(src)
			tmpSrc := filepath.Join(tmpDir, filepath.Base(src))
			os.WriteFile(tmpSrc, srcData, 0644)
			return run([]string{java, "-jar", jar, "-t" + format, tmpSrc}, tmpDir)
		}
	}
	return fmt.Errorf("PlantUML rendering requires `plantuml` or Java+PLANTUML_JAR")
}

func renderSVG(src, out, format string) error {
	if format == "svg" {
		data, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		return os.WriteFile(out, data, 0644)
	}
	rsvg, _ := exec.LookPath("rsvg-convert")
	if rsvg != "" {
		return run([]string{rsvg, "-f", format, "-o", out, src}, "")
	}
	inkscape, _ := exec.LookPath("inkscape")
	if inkscape != "" {
		return run([]string{inkscape, src, "--export-type=" + format, "--export-filename=" + out}, "")
	}
	return fmt.Errorf("SVG conversion requires `rsvg-convert` or `inkscape`")
}

func main() {
	input := flag.String("input", "", "Input diagram file")
	format := flag.String("format", "svg", "Output format: svg, png, pdf")
	outPath := flag.String("out", "", "Output path")
	kind := flag.String("kind", "auto", "Diagram kind: auto, mermaid, graphviz, plantuml, svg")
	flag.Parse()

	if *input == "" {
		if flag.NArg() > 0 {
			*input = flag.Arg(0)
		} else {
			fmt.Fprintln(os.Stderr, "Usage: render_diagram <input> [--format svg|png|pdf] [--out <path>]")
			os.Exit(2)
		}
	}

	src, err := filepath.Abs(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
		os.Exit(2)
	}

	if _, err := os.Stat(src); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "input file not found: %s\n", src)
		os.Exit(2)
	}

	out := *outPath
	if out == "" {
		out = strings.TrimSuffix(src, filepath.Ext(src)) + "." + *format
	}
	outDir := filepath.Dir(out)
	os.MkdirAll(outDir, 0755)

	k := inferKind(src)
	if *kind != "auto" {
		k = *kind
	}

	var renderErr error
	switch k {
	case "mermaid":
		renderErr = renderMermaid(src, out, *format)
	case "graphviz":
		renderErr = renderGraphviz(src, out, *format)
	case "plantuml":
		renderErr = renderPlantUML(src, out, *format)
	case "svg":
		renderErr = renderSVG(src, out, *format)
	default:
		renderErr = fmt.Errorf("unknown diagram kind: %s", k)
	}

	if renderErr != nil {
		fmt.Fprintf(os.Stderr, "render failed: %v\n", renderErr)
		os.Exit(1)
	}

	if info, err := os.Stat(out); err != nil || info.Size() == 0 {
		fmt.Fprintf(os.Stderr, "render failed: output missing or empty: %s\n", out)
		os.Exit(1)
	}
	fmt.Printf("rendered %s diagram to %s\n", k, out)
}

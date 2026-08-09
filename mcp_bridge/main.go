package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
)

const burpHost = "127.0.0.1"
const burpPort = "9876"

var burpBase string

var toolNames []string
var pending atomic.Int32
var stdinClosed atomic.Bool

func init() {
	if h := os.Getenv("BURP_MCP_HOST"); h != "" {
		_ = h // use const for simplicity
	}
	if p := os.Getenv("BURP_MCP_PORT"); p != "" {
		_ = p
	}
	burpBase = "http://" + burpHost + ":" + burpPort
}

func fetchTools() ([]string, error) {
	resp, err := http.Get(burpBase + "/tools")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var tools []string
	if err := json.Unmarshal(body, &tools); err != nil {
		return nil, err
	}
	return tools, nil
}

func callTool(toolName string, params map[string]interface{}) (interface{}, error) {
	body := map[string]interface{}{
		"tool":   toolName,
		"params": params,
	}
	data, _ := json.Marshal(body)
	resp, err := http.Post(burpBase+"/", "application/json", strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("Cannot reach Burp MCP at %s:%s (%v). Ensure Burp Suite is running with the \"MCP Full Control\" extension loaded.", burpHost, burpPort, err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var result interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return string(respBody), nil
	}
	return result, nil
}

func buildToolDefinitions() []map[string]interface{} {
	defs := make([]map[string]interface{}, len(toolNames))
	for i, name := range toolNames {
		defs[i] = map[string]interface{}{
			"name":        "burp_" + name,
			"description": getDescription(name),
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": getParams(name),
			},
		}
	}
	return defs
}

func getDescription(name string) string {
	descs := map[string]string{
		"proxy_history":           "Get Burp proxy history with optional filtering",
		"proxy_detail":            "Get full request/response details for a specific proxy history item",
		"proxy_websocket":         "Get WebSocket message history",
		"proxy_listeners":         "Get proxy listener information",
		"proxy_match_replace":     "Manage match & replace rules",
		"proxy_clear":             "Clear proxy history",
		"proxy_history_filtered":  "Filter proxy history by annotation color or notes",
		"send_request":            "Send an HTTP request through Burp and get the response",
		"send_to_repeater":        "Send a raw request to Burp Repeater tab",
		"repeater_send":           "Send a request and get response (like Repeater)",
		"repeater_modify_send":    "Modify headers/body of a request then send it",
		"send_to_intruder":        "Send a request to Burp Intruder",
		"sitemap":                 "Get site map entries with optional URL prefix filter",
		"target_info":             "Get target information",
		"intercept_toggle":        "Enable or disable proxy intercept",
		"encode":                  "Encode a string (base64, url, hex)",
		"decode":                  "Decode a string (base64, url)",
		"scan":                    "Start a vulnerability scan",
		"scan_active":             "Start active scan",
		"scan_results":            "Get scan results",
		"scan_issue_detail":       "Get detailed scan issue info",
		"crawl":                   "Start crawling a URL",
		"get_scope":               "Check if URL is in scope",
		"add_to_scope":            "Add URL to scope",
		"remove_from_scope":       "Remove URL from scope",
		"collaborator_generate":   "Generate Burp Collaborator payloads",
		"collaborator_poll":       "Poll for Collaborator interactions",
		"search_history":          "Search proxy history with regex",
		"highlight":               "Highlight a proxy history item",
		"annotate":                "Add a note to a proxy history item",
		"cookie_jar":              "View cookies in Burp cookie jar",
		"export_cert":             "Get instructions for exporting Burp CA certificate",
		"save_project":            "Save the current Burp project",
		"burp_version":            "Get Burp Suite version",
		"extensions_list":         "Get loaded extensions info",
		"log":                     "Write to Burp extension log",
		"add_issue":               "Manually add a vulnerability issue",
		"register_http_handler":   "Register auto-modify rule for HTTP requests",
		"remove_http_handler":     "Remove HTTP handler rules",
		"register_proxy_rule":     "Register proxy intercept rule",
		"remove_proxy_rule":       "Remove proxy intercept rules",
	}
	if d, ok := descs[name]; ok {
		return d
	}
	return "Burp Suite tool: " + name
}

func getParams(name string) map[string]interface{} {
	// Return empty schema for simplicity - Burp API handles validation
	return map[string]interface{}{}
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type toolsListParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

func handleRequest(msg rpcRequest) *rpcResponse {
	switch msg.Method {
	case "initialize":
		return &rpcResponse{
			JSONRPC: "2.0", ID: msg.ID,
			Result: map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities":    map[string]interface{}{"tools": map[string]interface{}{}},
				"serverInfo":      map[string]interface{}{"name": "burpsuite-mcp", "version": "2.0.0"},
			},
		}

	case "notifications/initialized":
		return nil

	case "tools/list":
		if len(toolNames) == 0 {
			return &rpcResponse{
				JSONRPC: "2.0", ID: msg.ID,
				Error: &rpcError{Code: -32000, Message: fmt.Sprintf("Burp MCP not connected at %s:%s.", burpHost, burpPort)},
			}
		}
		return &rpcResponse{
			JSONRPC: "2.0", ID: msg.ID,
			Result: map[string]interface{}{
				"tools": buildToolDefinitions(),
			},
		}

	case "tools/call":
		if len(toolNames) == 0 {
			return &rpcResponse{
				JSONRPC: "2.0", ID: msg.ID,
				Error: &rpcError{Code: -32000, Message: fmt.Sprintf("Burp MCP not connected at %s:%s.", burpHost, burpPort)},
			}
		}
		var params toolsListParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			return &rpcResponse{JSONRPC: "2.0", ID: msg.ID, Error: &rpcError{Code: -32602, Message: "Invalid params"}}
		}
		toolName := strings.TrimPrefix(params.Name, "burp_")
		result, err := callTool(toolName, params.Arguments)
		if err != nil {
			return &rpcResponse{JSONRPC: "2.0", ID: msg.ID, Error: &rpcError{Code: -1, Message: err.Error()}}
		}
		resData, _ := json.MarshalIndent(result, "", "  ")
		return &rpcResponse{
			JSONRPC: "2.0", ID: msg.ID,
			Result: map[string]interface{}{
				"content": []map[string]string{
					{"type": "text", "text": string(resData)},
				},
			},
		}

	default:
		return &rpcResponse{
			JSONRPC: "2.0", ID: msg.ID,
			Error: &rpcError{Code: -32601, Message: fmt.Sprintf("Method not found: %s", msg.Method)},
		}
	}
}

func main() {
	// Try to connect to Burp
	if tools, err := fetchTools(); err == nil {
		toolNames = tools
		fmt.Fprintf(os.Stderr, "[burp-mcp-bridge] Connected to Burp. %d tools available.\n", len(tools))
	} else {
		fmt.Fprintf(os.Stderr, "[burp-mcp-bridge] WARNING: Cannot connect to Burp at %s:%s. Start Burp first.\n", burpHost, burpPort)
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		pending.Add(1)

		var msg rpcRequest
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			fmt.Fprintln(os.Stderr, "Parse error:", err)
			resp, _ := json.Marshal(rpcResponse{
				JSONRPC: "2.0", ID: nil, Error: &rpcError{Code: -32700, Message: "Parse error"},
			})
			fmt.Println(string(resp))
			pending.Add(-1)
			continue
		}

		response := handleRequest(msg)
		if response != nil {
			respData, _ := json.Marshal(response)
			fmt.Println(string(respData))
		}
		pending.Add(-1)
	}

	stdinClosed.Store(true)
}

package mcpserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Response стандартный формат ответа
type Response struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Command string      `json:"command"`
}

type request struct {
	Command string          `json:"command"`
	Params  json.RawMessage `json:"params"`
}

// RunMCPServer запускает очень упрощённый MCP-сервер на указанном адресе.
// Поддерживается только команда chain_list, возвращающая пустой список.
func RunMCPServer(addr string) error {
	mux := http.NewServeMux()

	// инициализируем in-memory store
	store := newInMemoryStore()
	runs := newRunStore()

	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			send(w, Response{Status: "error", Error: fmt.Sprintf("invalid json: %v", err), Command: req.Command})
			return
		}

		switch req.Command {
		case "chain_list":
			chains := store.ListChains()
			data := map[string]interface{}{"chains": chains}
			send(w, Response{Status: "success", Data: data, Command: req.Command})

		case "chain_create":
			var p struct {
				Name        string      `json:"name"`
				Description string      `json:"description"`
				Steps       interface{} `json:"steps"`
			}
			if err := json.Unmarshal(req.Params, &p); err != nil {
				send(w, Response{Status: "error", Error: fmt.Sprintf("invalid params: %v", err), Command: req.Command})
				return
			}
			id := fmt.Sprintf("chain-%d", time.Now().UnixNano()/1e6)
			chain := Chain{
				ID:          id,
				Name:        p.Name,
				Description: p.Description,
				Steps:       p.Steps,
				CreatedAt:   time.Now(),
			}
			_, _ = store.CreateChain(chain)
			data := map[string]interface{}{"id": id}
			send(w, Response{Status: "success", Data: data, Command: req.Command})

		case "chain_get":
			var p struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(req.Params, &p); err != nil {
				send(w, Response{Status: "error", Error: fmt.Sprintf("invalid params: %v", err), Command: req.Command})
				return
			}
			if chain, ok := store.GetChain(p.ID); ok {
				send(w, Response{Status: "success", Data: chain, Command: req.Command})
			} else {
				send(w, Response{Status: "error", Error: "chain not found", Command: req.Command})
			}

		case "chain_delete":
			var p struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(req.Params, &p); err != nil {
				send(w, Response{Status: "error", Error: fmt.Sprintf("invalid params: %v", err), Command: req.Command})
				return
			}
			if ok := store.DeleteChain(p.ID); ok {
				send(w, Response{Status: "success", Data: map[string]interface{}{"id": p.ID}, Command: req.Command})
			} else {
				send(w, Response{Status: "error", Error: "chain not found", Command: req.Command})
			}

		case "chain_run":
			var p struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(req.Params, &p); err != nil {
				send(w, Response{Status: "error", Error: fmt.Sprintf("invalid params: %v", err), Command: req.Command})
				return
			}
			if _, ok := store.GetChain(p.ID); !ok {
				send(w, Response{Status: "error", Error: "chain not found", Command: req.Command})
				return
			}
			run := runs.createRun(p.ID)
			go runs.simulateRun(run.ID)
			send(w, Response{Status: "success", Data: run, Command: req.Command})

		case "chain_progress":
			var p struct {
				RunID string `json:"run_id"`
			}
			if err := json.Unmarshal(req.Params, &p); err != nil {
				send(w, Response{Status: "error", Error: fmt.Sprintf("invalid params: %v", err), Command: req.Command})
				return
			}
			if run, ok := runs.getRun(p.RunID); ok {
				send(w, Response{Status: "success", Data: run, Command: req.Command})
			} else {
				send(w, Response{Status: "error", Error: "run not found", Command: req.Command})
			}

		case "chain_results":
			var p struct {
				RunID string `json:"run_id"`
			}
			if err := json.Unmarshal(req.Params, &p); err != nil {
				send(w, Response{Status: "error", Error: fmt.Sprintf("invalid params: %v", err), Command: req.Command})
				return
			}
			if run, ok := runs.getRun(p.RunID); ok {
				if run.Status != RunStatusCompleted {
					send(w, Response{Status: "error", Error: "run not completed yet", Command: req.Command})
					return
				}
				send(w, Response{Status: "success", Data: map[string]interface{}{"result": run.Result}, Command: req.Command})
			} else {
				send(w, Response{Status: "error", Error: "run not found", Command: req.Command})
			}

		case "checkpoint_list":
			var p struct {
				RunID string `json:"run_id"`
			}
			if err := json.Unmarshal(req.Params, &p); err != nil {
				send(w, Response{Status: "error", Error: fmt.Sprintf("invalid params: %v", err), Command: req.Command})
				return
			}
			cps := globalCPStore.list(p.RunID)
			send(w, Response{Status: "success", Data: cps, Command: req.Command})

		case "checkpoint_get":
			var p struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(req.Params, &p); err != nil {
				send(w, Response{Status: "error", Error: fmt.Sprintf("invalid params: %v", err), Command: req.Command})
				return
			}
			if cp, ok := globalCPStore.get(p.ID); ok {
				send(w, Response{Status: "success", Data: cp, Command: req.Command})
			} else {
				send(w, Response{Status: "error", Error: "checkpoint not found", Command: req.Command})
			}

		default:
			send(w, Response{Status: "unknown_command", Error: "command not supported", Command: req.Command})
		}
	})

	log.Printf("MCP server listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}

func send(w http.ResponseWriter, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

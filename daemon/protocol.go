package main

import (
	"encoding/json"
)

// Request represents an incoming request from the CLI
type Request struct {
	Type    string          `json:"type"`
	ID      string          `json:"id"`
	Payload json.RawMessage `json:"payload"`
}

// Response represents the daemon's response
type Response struct {
	ID      string          `json:"id"`
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// Request types
const (
	RequestPossess = "possess"
	RequestList    = "list"
	RequestStatus  = "status"
	RequestMemory  = "memory"
	RequestEnd     = "end"
)

// PossessPayload for possession requests
type PossessPayload struct {
	Agent     string `json:"agent"`
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
}

// StatusData for status responses
type StatusData struct {
	Status   string `json:"status"`
	Port     string `json:"port"`
	Sessions int    `json:"sessions"`
	Uptime   string `json:"uptime"`
	Dolphins string `json:"dolphins"`
}

// ListData for list responses
type ListData struct {
	Commands []string `json:"commands"`
}

// Helper functions
func NewResponse(id string, success bool) Response {
	return Response{
		ID:      id,
		Success: success,
	}
}

func (r *Response) SetData(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	r.Data = jsonData
	return nil
}

func (r *Response) SetError(err string) {
	r.Success = false
	r.Error = err
}

// NewErrorResponse creates an error response
func NewErrorResponse(id string, errorMsg string) Response {
	resp := NewResponse(id, false)
	resp.SetError(errorMsg)
	return resp
}
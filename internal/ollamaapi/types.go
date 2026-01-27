package ollamaapi

import "time"

type VersionResponse struct {
	Version string `json:"version"`
}

type TagsResponse struct {
	Models []TagModel `json:"models"`
}

type TagModel struct {
	Name       string    `json:"name"`
	Digest     string    `json:"digest"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modified_at"`
}

type PSResponse struct {
	Models []PSModel `json:"models"`
}

type PSModel struct {
	Name       string    `json:"name"`
	Digest     string    `json:"digest"`
	Size       int64     `json:"size"`
	ExpiresAt  time.Time `json:"expires_at"`
	Details    any       `json:"details"`
	Model      string    `json:"model"`
	ModifiedAt time.Time `json:"modified_at"`
}

type ShowRequest struct {
	Name string `json:"name"`
}

type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type GenerateChunk struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error"`
}

type PullRequest struct {
	Name   string `json:"name"`
	Stream bool   `json:"stream"`
}

type PullChunk struct {
	Status    string `json:"status"`
	Digest    string `json:"digest"`
	Total     int64  `json:"total"`
	Completed int64  `json:"completed"`
	Error     string `json:"error"`
}

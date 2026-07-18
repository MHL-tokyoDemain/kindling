package parser

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/kindling/kindling/pkg/types"
)

type ParseResult struct {
	Document *types.Document
	Code     string
	Error    string
}

func Parse(filename string, content []byte, maxFileSize int64) ParseResult {
	if int64(len(content)) > maxFileSize {
		return ParseResult{
			Code:  types.ErrPayloadTooLarge,
			Error: "file exceeds maximum allowed size",
		}
	}

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		return parseJSON(filename, content)
	case ".txt":
		return parseText(filename, content)
	default:
		return ParseResult{
			Code:  types.ErrInvalidContentType,
			Error: "unsupported file extension: " + ext,
		}
	}
}

func parseJSON(filename string, content []byte) ParseResult {
	if len(content) == 0 {
		return ParseResult{
			Code:  types.ErrEmptyJSON,
			Error: "JSON file is empty",
		}
	}

	content = stripBOM(content)
	if len(content) == 0 {
		return ParseResult{
			Code:  types.ErrEmptyJSON,
			Error: "JSON file is empty",
		}
	}

	var data any
	if err := json.Unmarshal(content, &data); err != nil {
		return ParseResult{
			Code:  types.ErrParseFailed,
			Error: "invalid JSON: " + err.Error(),
		}
	}

	switch v := data.(type) {
	case map[string]any:
		return ParseResult{
			Document: &types.Document{
				Filename:    filename,
				Content:     v,
				ContentType: "json",
			},
		}
	case []any:
		return ParseResult{
			Code:  types.ErrParseFailed,
			Error: "expected JSON object, got array",
		}
	default:
		return ParseResult{
			Code:  types.ErrParseFailed,
			Error: "expected JSON object",
		}
	}
}

func parseText(filename string, content []byte) ParseResult {
	content = stripBOM(content)

	return ParseResult{
		Document: &types.Document{
			Filename:    filename,
			Content:     string(content),
			ContentType: "text",
		},
	}
}

func stripBOM(data []byte) []byte {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return data[3:]
	}
	return data
}

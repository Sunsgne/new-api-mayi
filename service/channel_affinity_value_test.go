package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestNormalizeChannelAffinityGJSONValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
		path string
		want string
	}{
		{
			name: "previous_response_id string",
			body: `{"previous_response_id":"resp_abc"}`,
			path: "previous_response_id",
			want: "resp_abc",
		},
		{
			name: "previous_response_id null",
			body: `{"previous_response_id":null}`,
			path: "previous_response_id",
			want: "",
		},
		{
			name: "previous_response_id empty",
			body: `{"previous_response_id":""}`,
			path: "previous_response_id",
			want: "",
		},
		{
			name: "file multipath without file_id yields empty",
			body: `{"input":[{"content":[{"type":"input_text","text":"hi"}]}]}`,
			path: "input.#.content.#.file_id",
			want: "",
		},
		{
			name: "file multipath extracts first file_id",
			body: `{"input":[{"content":[{"type":"input_file","file_id":"file-abc"},{"type":"input_text","text":"x"}]}]}`,
			path: "input.#.content.#.file_id",
			want: "file-abc",
		},
		{
			name: "code interpreter file_ids",
			body: `{"tools":[{"type":"code_interpreter","container":{"file_ids":["file-ci1","file-ci2"]}}]}`,
			path: "tools.0.container.file_ids.0",
			want: "file-ci1",
		},
		{
			name: "filtered content file_id",
			body: `{"input":[{"content":[{"type":"input_file","file_id":"file-xyz"}]}]}`,
			path: `input.0.content.#(file_id!="").file_id`,
			want: "file-xyz",
		},
		{
			name: "filtered content without file_id",
			body: `{"input":[{"content":[{"type":"input_text","text":"hi"}]}]}`,
			path: `input.0.content.#(file_id!="").file_id`,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := normalizeChannelAffinityGJSONValue(gjson.Get(tt.body, tt.path))
			assert.Equal(t, tt.want, got)
		})
	}
}

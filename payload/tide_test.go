package payload

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/wptide/pkg/message"
	"github.com/wptide/pkg/tide"
)

type MockTideClient struct {
	apiError bool
}

func (m MockTideClient) Authenticate(clientID, clientSecret, authEndpoint string) error {
	return nil
}

func (m MockTideClient) SendPayload(method, endpoint, data string) (string, error) {

	if endpoint == "http://test.local/fail" {
		return "", errors.New("something went wrong")
	}

	return "", nil
}

func TestTidePayload_BuildPayload(t *testing.T) {

	mockInfo := tide.CodeInfo{
		"plugin",
		[]tide.InfoDetails{},
		map[string]tide.ClocResult{},
	}

	type fields struct {
		Client tide.ClientInterface
	}
	type args struct {
		msg  message.Message
		data map[string]interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"No CodeInfo",
			fields{
				&MockTideClient{},
			},
			args{},
			nil,
			true,
		},
		{
			"No Results",
			fields{
				&MockTideClient{},
			},
			args{
				data: map[string]interface{}{
					"info": mockInfo,
				},
			},
			nil,
			true,
		},
		{
			"Some Results",
			fields{
				&MockTideClient{},
			},
			args{
				data: map[string]interface{}{
					"info": mockInfo,
					"phpcs_demo": tide.AuditResult{
						Raw: tide.AuditDetails{
							Type:     "mock",
							FileName: "mock",
							Path:     "mock",
						},
						Parsed: tide.AuditDetails{
							Type:     "mock",
							FileName: "mock",
							Path:     "mock",
						},
					},
					"checksum": "abcdefg",
				},
			},
			[]byte(`{"title":"","content":"","version":"","checksum":"abcdefg","visibility":"","project_type":"plugin","source_url":"","source_type":"","code_info":{"type":"plugin","details":[],"cloc":{}},"reports":{"phpcs_demo":{"raw":{"type":"mock","filename":"mock","path":"mock"},"parsed":{"type":"mock","filename":"mock","path":"mock"},"summary":{}}}}`),
			false,
		},
		{
			"Some Results - With Project Defined",
			fields{
				&MockTideClient{},
			},
			args{
				data: map[string]interface{}{
					"info": mockInfo,
					"phpcs_demo": tide.AuditResult{
						Raw: tide.AuditDetails{
							Type:     "mock",
							FileName: "mock",
							Path:     "mock",
						},
						Parsed: tide.AuditDetails{
							Type:     "mock",
							FileName: "mock",
							Path:     "mock",
						},
					},
					"checksum": "abcdefg",
				},
				msg: message.Message{
					Slug: "project-one",
				},
			},
			[]byte(`{"title":"","content":"","version":"","checksum":"abcdefg","visibility":"","project_type":"plugin","source_url":"","source_type":"","code_info":{"type":"plugin","details":[],"cloc":{}},"reports":{"phpcs_demo":{"raw":{"type":"mock","filename":"mock","path":"mock"},"parsed":{"type":"mock","filename":"mock","path":"mock"},"summary":{}}},"project":["project-one"]}`),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := TidePayload{
				Client: tt.fields.Client,
			}
			got, err := tp.BuildPayload(tt.args.msg, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("TidePayload.BuildPayload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				fmt.Println(string(got))
				t.Errorf("TidePayload.BuildPayload() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func Test_fallbackValue(t *testing.T) {
	type args struct {
		value []interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			"Mismatch First Item",
			args{
				[]interface{}{
					"one",
					2,
				},
			},
			"one",
		},
		{
			"Mismatch Second Item",
			args{
				[]interface{}{
					"",
					2,
				},
			},
			2,
		},
		{
			"Strings",
			args{
				[]interface{}{
					"",
					"",
					"three",
				},
			},
			"three",
		},
		{
			"Strings Empty",
			args{
				[]interface{}{
					"",
					"",
				},
			},
			"",
		},
		{
			"int64",
			args{
				[]interface{}{
					int64(4),
					int64(2),
				},
			},
			int64(4),
		},
		{
			"int64 Empty",
			args{
				[]interface{}{
					int64(0),
					int64(0),
				},
			},
			int64(0),
		},
		{
			"int32",
			args{
				[]interface{}{
					int32(0),
					int32(12),
				},
			},
			int32(12),
		},
		{
			"int32 Empty",
			args{
				[]interface{}{
					int32(0),
					int32(0),
				},
			},
			int32(0),
		},
		{
			"int",
			args{
				[]interface{}{
					0,
					0,
					42,
				},
			},
			42,
		},
		{
			"int Empty",
			args{
				[]interface{}{
					0,
					0,
				},
			},
			0,
		},
		{
			"float64",
			args{
				[]interface{}{
					float64(0.0),
					float64(42.0),
					float64(0.0),
				},
			},
			float64(42.0),
		},
		{
			"float64 Empty",
			args{
				[]interface{}{
					float64(0.0),
					float64(0.0),
				},
			},
			float64(0.0),
		},
		{
			"float32",
			args{
				[]interface{}{
					float32(0.0),
					float32(42.0),
					float32(0.0),
				},
			},
			float32(42.0),
		},
		{
			"float32 Empty",
			args{
				[]interface{}{
					float32(0.0),
					float32(0.0),
				},
			},
			float32(0.0),
		},
		{
			"Other - default",
			args{
				[]interface{}{
					[]string{"a"},
					[]string{"b"},
				},
			},
			[]string{"a"},
		},
		{
			"CodeInfo",
			args{
				[]interface{}{
					tide.CodeInfo{Type: ""},
					tide.CodeInfo{Type: "plugin"},
				},
			},
			tide.CodeInfo{Type: "plugin"},
		},
		{
			"CodeInfo Empty",
			args{
				[]interface{}{
					tide.CodeInfo{Type: ""},
					tide.CodeInfo{Type: ""},
				},
			},
			nil,
		},
		{
			"No args",
			args{},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fallbackValue(tt.args.value...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fallbackValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTidePayload_SendPayload(t *testing.T) {
	type fields struct {
		Client tide.ClientInterface
	}
	type args struct {
		destination string
		payload     []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"Successful Send",
			fields{
				&MockTideClient{},
			},
			args{
				"http://test.local/endpoint",
				[]byte(`{"some":"payload"}`),
			},
			[]byte(""),
			false,
		},
		{
			"Failed Send",
			fields{
				&MockTideClient{},
			},
			args{
				"http://test.local/fail",
				[]byte(`{"failed":"payload"}`),
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := TidePayload{
				Client: tt.fields.Client,
			}
			got, err := tp.SendPayload(tt.args.destination, tt.args.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("TidePayload.SendPayload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TidePayload.SendPayload() = %v, want %v", got, tt.want)
			}
		})
	}
}

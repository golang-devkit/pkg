package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"sort"
	"strings"
)

// MultipartFile describes one multipart upload field.
type MultipartFile struct {
	FieldName   string
	FileName    string
	ContentType string
	Open        func() (io.ReadCloser, error)
}

// NewJSONRequest builds a JSON POST request.
func NewJSONRequest(ctx context.Context, endpoint string, body any) (*http.Request, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, fmt.Errorf("encode JSON request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// NewMultipartRequest builds a streaming multipart POST request.
func NewMultipartRequest(ctx context.Context, endpoint string, fields map[string]string, files []MultipartFile) (*http.Request, error) {
	reader, writer := io.Pipe()
	mpw := multipart.NewWriter(writer)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mpw.FormDataContentType())

	go func() {
		var buildErr error
		defer func() {
			closeErr := mpw.Close()
			if buildErr == nil && closeErr != nil {
				buildErr = closeErr
			}
			_ = writer.CloseWithError(buildErr)
		}()

		keys := make([]string, 0, len(fields))
		for key := range fields {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			if buildErr = mpw.WriteField(key, fields[key]); buildErr != nil {
				return
			}
		}

		for _, file := range files {
			if file.Open == nil {
				continue
			}
			rc, err := file.Open()
			if err != nil {
				buildErr = err
				return
			}

			part, err := createPart(mpw, file)
			if err != nil {
				_ = rc.Close()
				buildErr = err
				return
			}

			_, err = io.Copy(part, rc)
			closeErr := rc.Close()
			if err != nil {
				buildErr = err
				return
			}
			if closeErr != nil {
				buildErr = closeErr
				return
			}
		}
	}()

	return req, nil
}

func createPart(mpw *multipart.Writer, file MultipartFile) (io.Writer, error) {
	if file.ContentType == "" {
		return mpw.CreateFormFile(file.FieldName, file.FileName)
	}

	header := make(textproto.MIMEHeader)
	header.Set(
		"Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`, escapeQuotes(file.FieldName), escapeQuotes(file.FileName)),
	)
	header.Set("Content-Type", file.ContentType)
	return mpw.CreatePart(header)
}

func escapeQuotes(value string) string {
	return strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(value)
}

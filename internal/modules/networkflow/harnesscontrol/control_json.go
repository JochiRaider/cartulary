package harnesscontrol

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
)

func decodeControlRequest(r *http.Request, request any, required ...string) error {
	if r.Body == nil {
		return errors.New("body is required")
	}
	defer r.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(r.Body, (1<<20)+1))
	if err != nil {
		return err
	}
	if len(raw) > 1<<20 {
		return errors.New("control request too large")
	}
	object, err := strictjson.DecodeObject(bytes.NewReader(raw))
	if err != nil {
		return errors.New("body must be one strict JSON object")
	}
	for _, key := range required {
		if _, ok := object[key]; !ok {
			return errors.New("required control field missing")
		}
	}
	for _, value := range object {
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return errors.New("control fields must not be null")
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	return decoder.Decode(request)
}

package message

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrInvalidPayload = errors.New("invalid payload")
	ErrInvalidData    = errors.New("invalid data")
)

type Payload interface {
	Type() string
}

type TypedData struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func UnmarshalType(data []byte) (TypedData, error) {
	var typedData TypedData

	if len(data) == 0 {
		return typedData, fmt.Errorf("%w: data is nil or empty", ErrInvalidData)
	}

	if err := json.Unmarshal(data, &typedData); err != nil {
		return typedData, fmt.Errorf("%w: unexpected data format, %s: %s", ErrInvalidData, data, err)
	}

	return typedData, nil
}

func Unmarshal[T Payload](typedData TypedData, payload *T) error {
	if len(typedData.Data) == 0 {
		return fmt.Errorf("%w: data is nil or empty", ErrInvalidData)
	}

	if payload == nil {
		return fmt.Errorf("%w: payload is nil", ErrInvalidPayload)
	}

	payloadType := (*payload).Type()
	if typedData.Type != payloadType {
		return fmt.Errorf("%w: expected type '%s', got '%s'", ErrInvalidPayload, payloadType, typedData.Type)
	}

	if err := json.Unmarshal(typedData.Data, payload); err != nil {
		return fmt.Errorf("%w: invalid data for payload type '%s', %s: %s", ErrInvalidData, payloadType, typedData.Data, err)
	}

	return nil
}

func Marshal[T Payload](payload T) ([]byte, error) {
	type Message struct {
		Type string `json:"type"`
		Data any    `json:"data"`
	}

	payloadBytes, err := json.Marshal(Message{
		Type: payload.Type(),
		Data: payload,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: could not marshal, %s", ErrInvalidPayload, err)
	}

	return payloadBytes, nil
}

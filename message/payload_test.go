package message_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wspowell/tabletop/game"
	"github.com/wspowell/tabletop/message"
)

func removeWhitespace(input []byte) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(string(input), " ", ""), "\n", ""), "\t", "")
}

func TestUnmarshalType_nil_data(t *testing.T) {
	t.Parallel()

	_, err := message.UnmarshalType(nil)
	require.ErrorIs(t, err, message.ErrInvalidData)
}

func TestUnmarshalType_malformed_data(t *testing.T) {
	t.Parallel()

	_, err := message.UnmarshalType([]byte(`{`))
	require.ErrorIs(t, err, message.ErrInvalidData)
}

func TestUnmarshal_nil_data(t *testing.T) {
	t.Parallel()

	var payload message.Login
	err := message.Unmarshal(message.TypedData{
		Type: "login",
		Data: nil,
	}, &payload)
	require.ErrorIs(t, err, message.ErrInvalidData)
}

func TestUnmarshal_empty_data(t *testing.T) {
	t.Parallel()

	var payload message.Login
	err := message.Unmarshal(message.TypedData{
		Type: "login",
		Data: []byte(``),
	}, &payload)
	require.ErrorIs(t, err, message.ErrInvalidData)
}

func TestUnmarshal_nil_payload(t *testing.T) {
	t.Parallel()

	err := message.Unmarshal(message.TypedData{
		Type: "login",
		Data: []byte(`{}`),
	}, (*message.Login)(nil))
	require.ErrorIs(t, err, message.ErrInvalidPayload)
}

func TestUnmarshal_malformed_data(t *testing.T) {
	t.Parallel()

	var payload message.Login
	err := message.Unmarshal(message.TypedData{
		Type: "login",
		Data: []byte(`{`),
	}, &payload)
	require.ErrorIs(t, err, message.ErrInvalidData)
}

func TestUnmarshal_invalid_payload_type(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"type": "invalid",
		"data": {}
	}`)

	typedData, err := message.UnmarshalType(data)
	require.NoError(t, err)

	var payload message.Login
	err = message.Unmarshal(typedData, &payload)
	require.ErrorIs(t, err, message.ErrInvalidPayload)
}

func TestUnmarshal_login(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"type": "login",
		"data": {
			"username": "Foo",
			"secret":"Bar"
		}
	}`)

	typedData, err := message.UnmarshalType(data)
	require.NoError(t, err)

	var payload message.Login
	err = message.Unmarshal(typedData, &payload)
	require.NoError(t, err)
	assert.Equal(t, message.Login{
		Username: "Foo",
		Secret:   "Bar",
	}, payload)

	payloadBytes, err := message.Marshal(payload)
	require.NoError(t, err)
	assert.Equal(t, removeWhitespace(data), removeWhitespace(payloadBytes))
}

func TestUnmarshal_tokenPosition(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"type": "tokenPosition",
		"data": {
			"id": "token_1",
			"tokenName": "token",
			"mapName": "map",
			"position": {
				"x": 123,
				"y": 456
			}
		}
	}`)

	typedData, err := message.UnmarshalType(data)
	require.NoError(t, err)

	var payload message.TokenPosition
	err = message.Unmarshal(typedData, &payload)
	require.NoError(t, err)
	assert.Equal(t, message.TokenPosition{
		Id:        "token_1",
		TokenName: "token",
		MapName:   "map",
		Position: game.Coordinate{
			X: 123,
			Y: 456,
		},
		IsHome: false,
	}, payload)

	payloadBytes, err := message.Marshal(payload)
	require.NoError(t, err)
	assert.Equal(t, removeWhitespace(data), removeWhitespace(payloadBytes))
}

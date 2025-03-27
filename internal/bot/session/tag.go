package session

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log/slog"

	"github.com/CGSG-2021-AE4/tomestobot/api"
	"github.com/google/uuid"
)

// Stores value with a tag to verify that this value is the one you are supposed to get

// Tag type
type Tag uuid.UUID

var TagNil = Tag(uuid.Nil)

const TagLen = 16

// From/To bytes conversion

func TagFromBytes(b []byte) (Tag, error) {
	id, err := uuid.FromBytes(b)
	return Tag(id), err
}

func (tag Tag) Bytes() [16]byte {
	return [16]byte(uuid.UUID(tag))
}

// Main struct
type taggedVar[T any] struct {
	data T
	tag  Tag
}

type TaggedVar[T any] interface {
	Get(tag Tag) (T, error) // Returns data if only the tag is correct
	Set(data T) Tag         // Stores value and returns current tag
	// Tag() Tag // Just returns current tag
}

func newTaggedVar[T any]() TaggedVar[T] {
	return &taggedVar[T]{}
}

func (v *taggedVar[T]) Get(tag Tag) (T, error) {
	if v.tag != tag {
		return v.data, api.ErrorInvalidTag // Return data anyway because otherwise will have to create it
	}
	return v.data, nil
}

func (v *taggedVar[T]) Set(data T) Tag {
	v.data = data
	v.tag = Tag(uuid.New())
	return v.tag
}

// Some supplement functions because main usage - pack/unpacking to payload string

// Decodes Tag from bytes
func decodeTag(data string) (Tag, error) {
	bytes, err := hex.DecodeString(data)
	if err != nil {
		slog.Debug("hex decode err")
		return TagNil, api.ErrorInvalidBtnPayload
	}
	if len(bytes) != 16 { // uuid(16)
		slog.Debug("invalid btn payload len")
		return TagNil, api.ErrorInvalidBtnPayload
	}
	tag, err := TagFromBytes([]byte(bytes)[:16])
	if err != nil {
		slog.Debug("err while decode tag from bytes")
		return TagNil, api.ErrorInvalidBtnPayload
	}
	return tag, nil
}

// Decodes tag and a number after if
func decodeTagWithI(data string) (Tag, int, error) {
	bytes, err := hex.DecodeString(data)
	if err != nil {
		slog.Debug("hex decode err")
		return TagNil, 0, api.ErrorInvalidBtnPayload
	}
	slog.Debug(fmt.Sprint(bytes))
	if len(bytes) != 20 { // uuid(16) + uint32(4)
		slog.Debug("invalid btn payload len")
		return TagNil, 0, api.ErrorInvalidBtnPayload
	}
	tag, err := TagFromBytes([]byte(bytes)[:16])
	if err != nil {
		slog.Debug("err while decode tag from bytes")
		return TagNil, 0, api.ErrorInvalidBtnPayload
	}
	i := binary.LittleEndian.Uint32(bytes[16:])
	return tag, int(i), nil
}

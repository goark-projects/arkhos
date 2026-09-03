package hertz

import (
	"bytes"
	"strings"

	"github.com/cloudwego/hertz/pkg/protocol"
)

type requestHeader struct {
	header *protocol.RequestHeader
}

func (h requestHeader) Get(name string) string {
	return string(h.header.Peek(name))
}

func (h requestHeader) Values(name string) []string {
	values := h.header.PeekAll(name)
	result := make([]string, len(values))
	for i := range values {
		result[i] = string(values[i])
	}
	return result
}

func (h requestHeader) Has(name string) bool {
	found := false
	h.header.VisitAll(func(key, _ []byte) {
		found = found || bytes.EqualFold(key, []byte(name))
	})
	return found
}

func (h requestHeader) Set(name, value string) {
	h.header.Set(name, value)
}

func (h requestHeader) Add(name, value string) {
	h.header.Add(name, value)
}

func (h requestHeader) Delete(name string) {
	h.header.Del(name)
}

func (h requestHeader) Visit(visitor func(name, value string) bool) {
	if visitor == nil {
		return
	}
	active := true
	h.header.VisitAll(func(key, value []byte) {
		if active {
			active = visitor(string(key), string(value))
		}
	})
}

type responseHeader struct {
	header *protocol.ResponseHeader
}

func (h responseHeader) Get(name string) string {
	return string(h.header.Peek(name))
}

func (h responseHeader) Values(name string) []string {
	values := h.header.PeekAll(name)
	result := make([]string, len(values))
	for i := range values {
		result[i] = string(values[i])
	}
	return result
}

func (h responseHeader) Has(name string) bool {
	found := false
	h.header.VisitAll(func(key, _ []byte) {
		found = found || strings.EqualFold(string(key), name)
	})
	return found
}

func (h responseHeader) Set(name, value string) {
	h.header.Set(name, value)
}

func (h responseHeader) Add(name, value string) {
	h.header.Add(name, value)
}

func (h responseHeader) Delete(name string) {
	h.header.Del(name)
}

func (h responseHeader) Visit(visitor func(name, value string) bool) {
	if visitor == nil {
		return
	}
	active := true
	h.header.VisitAll(func(key, value []byte) {
		if active {
			active = visitor(string(key), string(value))
		}
	})
}

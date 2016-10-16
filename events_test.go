package main

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestProcessMessage(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name               string
		m                  message
		rs                 []rule
		fa                 map[string]string
		ie                 []event
		globalFSMId        string
		expectedEvents     []event
		expectedIncomplete []event
		expectedFa         map[string]string
	}{
		{
			name:           "empty case",
			m:              message{},
			rs:             []rule{},
			fa:             map[string]string{},
			expectedEvents: []event{},
			expectedFa:     map[string]string{},
		},
		{
			name: "matching topic name",
			m: message{
				Key:       "123",
				Value:     newValueFrom("{}"),
				Topic:     "topic",
				Partition: 0,
				Offset:    213,
				Timestamp: now,
			},
			rs: []rule{
				{
					Patterns: []pattern{{Field: "{{.Topic}}", Pattern: "topic"}},
					Events:   []event{{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi!", FSMId: "456"}},
				},
			},
			fa: map[string]string{},
			expectedEvents: []event{
				{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi!", FSMId: "456", JSON: newSliceFrom("{}")},
			},
			expectedFa: map[string]string{},
		},
		{
			name: "matching topic name and producing 2 events",
			m: message{
				Key:       "123",
				Value:     newValueFrom("{}"),
				Topic:     "topic",
				Partition: 0,
				Offset:    213,
				Timestamp: now,
			},
			rs: []rule{
				{
					Patterns: []pattern{{Field: "{{.Topic}}", Pattern: "topic"}},
					Events: []event{
						{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi, B!", FSMId: "456"},
						{EventType: "message", SourceId: "A", TargetId: "C", Text: "Hi, C!", FSMId: "456"},
					},
				},
			},
			fa: map[string]string{},
			expectedEvents: []event{
				{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi, B!", FSMId: "456", JSON: newSliceFrom("{}")},
				{EventType: "message", SourceId: "A", TargetId: "C", Text: "Hi, C!", FSMId: "456", JSON: newSliceFrom("{}")},
			},
			expectedFa: map[string]string{},
		},
		{
			name: "matching topic name and key using regex stuff for key",
			m: message{
				Key:       "123",
				Value:     newValueFrom("{}"),
				Topic:     "topic",
				Partition: 0,
				Offset:    213,
				Timestamp: now,
			},
			rs: []rule{
				{
					Patterns: []pattern{{Field: "{{.Topic}}", Pattern: "topic"}, {Field: "{{.Key}}", Pattern: `\d+`}},
					Events:   []event{{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi!", FSMId: "456"}},
				},
			},
			fa:             map[string]string{},
			expectedEvents: []event{{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi!", FSMId: "456", JSON: newSliceFrom("{}")}},
			expectedFa:     map[string]string{},
		},
		{
			name: "matching topic name and mismatching key regex",
			m: message{
				Key:       "123",
				Value:     newValueFrom("{}"),
				Topic:     "topic",
				Partition: 0,
				Offset:    213,
				Timestamp: now,
			},
			rs: []rule{
				{
					Patterns: []pattern{{Field: "{{.Topic}}", Pattern: "topic"}, {Field: "{{.Key}}", Pattern: `\d+not number`}},
					Events:   []event{{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi!", FSMId: "456", JSON: newSliceFrom("{}")}},
				},
			},
			fa:             map[string]string{},
			expectedEvents: []event{},
			expectedFa:     map[string]string{},
		},
		{
			name: "matching value",
			m: message{
				Key:       "123",
				Value:     newValueFrom(`{"name":"relevant"}`),
				Topic:     "topic",
				Partition: 0,
				Offset:    213,
				Timestamp: now,
			},
			rs: []rule{
				{
					Patterns: []pattern{{Field: `{{index .Value "name"}}`, Pattern: "relevant"}},
					Events:   []event{{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi!", FSMId: "456"}},
				},
			},
			fa:             map[string]string{},
			expectedEvents: []event{{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi!", FSMId: "456", JSON: newSliceFrom(`{"name":"relevant"}`)}},
			expectedFa:     map[string]string{},
		},
		{
			name: "matching int value; replacing text with content of specific value key",
			m: message{
				Key:       "123",
				Value:     newValueFrom(`{"value":666}`),
				Topic:     "topic",
				Partition: 0,
				Offset:    213,
				Timestamp: now,
			},
			rs: []rule{
				{
					Patterns: []pattern{{Field: `{{index .Value "value"}}`, Pattern: `\d+`}},
					Events:   []event{{EventType: "message", SourceId: "A", TargetId: "B", Text: `{{index .Value "value"}}`, FSMId: "456"}},
				},
			},
			fa:             map[string]string{},
			expectedEvents: []event{{EventType: "message", SourceId: "A", TargetId: "B", Text: "666", FSMId: "456", JSON: newSliceFrom(`{"value":666}`)}},
			expectedFa:     map[string]string{},
		},
		{
			name: "adding keyalias",
			m: message{
				Key:       "123",
				Value:     newValueFrom(`{"primary":666, "secondary":777}`),
				Topic:     "topic",
				Partition: 0,
				Offset:    213,
				Timestamp: now,
			},
			rs: []rule{
				{
					Patterns: []pattern{{Field: `{{index .Value "primary"}}`, Pattern: `\d+`}},
					Events: []event{
						{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi", FSMId: `{{index .Value "primary"}}`, FSMIdAlias: `{{index .Value "secondary"}}`},
					},
				},
			},
			fa: map[string]string{},
			expectedEvents: []event{
				{EventType: "alias", FSMId: "666", FSMIdAlias: "777"},
				{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi", FSMId: "666", JSON: newSliceFrom(`{"primary":666, "secondary":777}`)}},
			expectedFa: map[string]string{"777": "666"},
		},
		{
			name: "using keyalias to produce event with an original key not present in the payload",
			m: message{
				Key:       "123",
				Value:     newValueFrom(`{"secondary":777}`),
				Topic:     "topic",
				Partition: 0,
				Offset:    213,
				Timestamp: now,
			},
			rs: []rule{
				{
					Patterns: []pattern{{Field: `{{.Topic}}`, Pattern: `topic`}},
					Events: []event{
						{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi", FSMId: `{{index .Value "secondary"}}`},
					},
				},
			},
			fa: map[string]string{"777": "666"},
			expectedEvents: []event{
				{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi", FSMId: "666", JSON: newSliceFrom(`{"secondary":777}`)},
			},
			expectedFa: map[string]string{"777": "666"},
		},
		{
			name: "ignores message when global key doesn't match",
			m: message{
				Key:       "123",
				Value:     newValueFrom("{}"),
				Topic:     "topic",
				Partition: 0,
				Offset:    213,
				Timestamp: now,
			},
			rs: []rule{
				{
					Patterns: []pattern{{Field: "{{.Topic}}", Pattern: "topic"}},
					Events:   []event{{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi!", FSMId: "456"}},
				},
			},
			globalFSMId:    "789",
			fa:             map[string]string{},
			expectedEvents: []event{},
			expectedFa:     map[string]string{},
		},
		{
			name: "doesn't ignore message when global key matches",
			m: message{
				Key:       "123",
				Value:     newValueFrom("{}"),
				Topic:     "topic",
				Partition: 0,
				Offset:    213,
				Timestamp: now,
			},
			rs: []rule{
				{
					Patterns: []pattern{{Field: "{{.Topic}}", Pattern: "topic"}},
					Events:   []event{{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi!", FSMId: "456"}},
				},
			},
			globalFSMId: "456",
			fa:          map[string]string{},
			expectedEvents: []event{
				{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi!", FSMId: "456", JSON: newSliceFrom("{}")},
			},
			expectedFa: map[string]string{},
		},
		{
			name: "processes an incomplete event with no fsmId when the association appears (with globalFSMId filter)",
			m: message{
				Key:       "123",
				Value:     newValueFrom("{}"),
				Topic:     "topic",
				Partition: 0,
				Offset:    213,
				Timestamp: now,
			},
			rs: []rule{
				{
					Patterns: []pattern{{Field: "{{.Topic}}", Pattern: "topic"}},
					Events:   []event{{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi!", FSMId: "456", FSMIdAlias: "789"}},
				},
			},
			ie:          []event{{EventType: "message", SourceId: "B", TargetId: "C", Text: "Hi!", FSMIdAlias: "789", JSON: newSliceFrom("{}")}},
			globalFSMId: "456",
			fa:          map[string]string{},
			expectedEvents: []event{
				{EventType: "alias", FSMId: "456", FSMIdAlias: "789"},
				{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi!", FSMId: "456", JSON: newSliceFrom("{}")},
			},
			expectedIncomplete: []event{
				{EventType: "message", SourceId: "B", TargetId: "C", Text: "Hi!", FSMId: "456", JSON: newSliceFrom("{}")},
			},
			expectedFa: map[string]string{"789": "456"},
		},
		{
			name: "ignores an incomplete event with no fsmId when the association appears, due to globalFSMId filter",
			m: message{
				Key:       "123",
				Value:     newValueFrom("{}"),
				Topic:     "topic",
				Partition: 0,
				Offset:    213,
				Timestamp: now,
			},
			rs: []rule{
				{
					Patterns: []pattern{{Field: "{{.Topic}}", Pattern: "topic"}},
					Events:   []event{{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi!", FSMId: "456", FSMIdAlias: "789"}},
				},
			},
			ie:          []event{{EventType: "message", SourceId: "B", TargetId: "C", Text: "Hi!", FSMIdAlias: "789", JSON: newSliceFrom("{}")}},
			globalFSMId: "345",
			fa:          map[string]string{},
			expectedEvents: []event{
				{EventType: "alias", FSMId: "456", FSMIdAlias: "789"},
			},
			expectedIncomplete: []event{{EventType: "message", SourceId: "B", TargetId: "C", Text: "Hi!", FSMId: "456", JSON: newSliceFrom("{}")}},
			expectedFa:         map[string]string{"789": "456"},
		},
	}

	for _, ts := range tests {
		actualEvents := []event{}
		err := processMessage(ts.m, ts.rs, ts.fa, &actualEvents, &ts.ie, ts.globalFSMId)

		if err != nil {
			t.Errorf("'%v' shouldn't have failed, but did with %v", ts.name, err)
		}
		if !reflect.DeepEqual(actualEvents, ts.expectedEvents) {
			t.Errorf("on '%v': events mismatch; expected %+v but got %+v", ts.name, ts.expectedEvents, actualEvents)
		}
		if !reflect.DeepEqual(ts.fa, ts.expectedFa) {
			t.Errorf("on '%v': fsmId aliases not updated; expected %+v but got %+v", ts.name, ts.expectedFa, ts.fa)
		}
		if !reflect.DeepEqual(ts.ie, ts.expectedIncomplete) {
			t.Errorf("on '%v': incomplete events mismatch; expected %+v but got %+v", ts.name, ts.expectedIncomplete, ts.ie)
		}
	}
}

func newValueFrom(j string) map[string]interface{} {
	var v interface{}
	json.Unmarshal([]byte(j), &v)
	return v.(map[string]interface{})
}

func newSliceFrom(j string) []map[string]interface{} {
	return []map[string]interface{}{newValueFrom(j)}
}

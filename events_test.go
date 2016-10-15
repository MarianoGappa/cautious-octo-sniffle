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
		name           string
		m              message
		rs             []rule
		fa             map[string]string
		globalFSMId    string
		expectedEvents []event
		expectedKa     map[string]string
		fails          bool
	}{
		{
			name:           "empty case",
			m:              message{},
			rs:             []rule{},
			fa:             map[string]string{},
			expectedEvents: []event{},
			expectedKa:     map[string]string{},
			fails:          false,
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
				{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi!", FSMId: "456", Key: "123", JSON: map[string]interface{}{}},
			},
			expectedKa: map[string]string{},
			fails:      false,
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
				{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi, B!", FSMId: "456", Key: "123", JSON: map[string]interface{}{}},
				{EventType: "message", SourceId: "A", TargetId: "C", Text: "Hi, C!", FSMId: "456", Key: "123", JSON: map[string]interface{}{}},
			},
			expectedKa: map[string]string{},
			fails:      false,
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
			expectedEvents: []event{{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi!", FSMId: "456", Key: "123", JSON: map[string]interface{}{}}},
			expectedKa:     map[string]string{},
			fails:          false,
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
					Events:   []event{{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi!", FSMId: "456", Key: "123", JSON: map[string]interface{}{}}},
				},
			},
			fa:             map[string]string{},
			expectedEvents: []event{},
			expectedKa:     map[string]string{},
			fails:          false,
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
			expectedEvents: []event{{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi!", FSMId: "456", Key: "123", JSON: newValueFrom(`{"name":"relevant"}`)}},
			expectedKa:     map[string]string{},
			fails:          false,
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
			expectedEvents: []event{{EventType: "message", SourceId: "A", TargetId: "B", Text: "666", FSMId: "456", Key: "123", JSON: newValueFrom(`{"value":666}`)}},
			expectedKa:     map[string]string{},
			fails:          false,
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
			fa:             map[string]string{},
			expectedEvents: []event{{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi", FSMId: "666", Key: "123", JSON: newValueFrom(`{"primary":666, "secondary":777}`)}},
			expectedKa:     map[string]string{"777": "666"},
			fails:          false,
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
			fa:             map[string]string{"777": "666"},
			expectedEvents: []event{{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi", FSMId: "666", Key: "123", JSON: newValueFrom(`{"secondary":777}`)}},
			expectedKa:     map[string]string{"777": "666"},
			fails:          false,
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
			expectedKa:     map[string]string{},
			fails:          false,
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
				{EventType: "message", SourceId: "A", TargetId: "B", Text: "Hi!", FSMId: "456", Key: "123", JSON: map[string]interface{}{}},
			},
			expectedKa: map[string]string{},
			fails:      false,
		},
	}

	for _, ts := range tests {
		actualEvents, err := processMessage(ts.m, ts.rs, ts.fa, ts.globalFSMId)
		if ts.fails && err == nil {
			t.Errorf("'%v' should have failed", ts.name)
			t.FailNow()
		}

		if err != nil {
			t.Errorf("'%v' shouldn't have failed, but did with %v", ts.name, err)
		}
		if !reflect.DeepEqual(actualEvents, ts.expectedEvents) {
			t.Errorf("on '%v': events mismatch; expected %+v but got %+v", ts.name, ts.expectedEvents, actualEvents)
		}
		if !reflect.DeepEqual(ts.fa, ts.expectedKa) {
			t.Errorf("on '%v': key aliases not updated; expected %+v but got %+v", ts.name, ts.expectedKa, ts.fa)
		}
	}
}

func newValueFrom(j string) map[string]interface{} {
	var v interface{}
	json.Unmarshal([]byte(j), &v)
	return v.(map[string]interface{})
}

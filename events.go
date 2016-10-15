package main

import (
	"bytes"
	"regexp"
	"text/template"
)

func processMessage(m message, rules []rule, fsmIdAliases map[string]string, globalFSMId string) ([]event, error) {
	events := []event{}
	for _, r := range rules {
		pass := true
		for _, p := range r.Patterns {
			b, err := parseTempl(p.Field, m)
			if err != nil {
				return []event{}, err
			}

			matched, err := regexp.Match(p.Pattern, b)
			if err != nil {
				return []event{}, err
			}

			if !matched {
				pass = false
				break
			}
		}
		if !pass {
			continue
		}
		for _, e := range r.Events {
			bEventType, err := parseTempl(e.EventType, m)
			if err != nil {
				return events, err
			}
			bFSMId, err := parseTempl(e.FSMId, m)
			if err != nil {
				return events, err
			}
			bFSMIdAlias, err := parseTempl(e.FSMIdAlias, m)
			if err != nil {
				return events, err
			}
			bSourceId, err := parseTempl(e.SourceId, m)
			if err != nil {
				return events, err
			}
			bTargetId, err := parseTempl(e.TargetId, m)
			if err != nil {
				return events, err
			}
			bText, err := parseTempl(e.Text, m)
			if err != nil {
				return events, err
			}

			fsmId := string(bFSMId)
			if fa, ok := fsmIdAliases[fsmId]; len(fa) > 0 && ok {
				fsmId = fa
			}
			if len(bFSMIdAlias) > 0 && len(bFSMId) > 0 {
				fsmIdAliases[string(bFSMIdAlias)] = fsmId
			}

			if len(globalFSMId) > 0 && globalFSMId != fsmId {
				continue
			}

			events = append(events, event{
				EventType: string(bEventType),
				FSMId:     fsmId,
				SourceId:  string(bSourceId),
				TargetId:  string(bTargetId),
				Text:      string(bText),
				JSON:      m.Value,
				Key:       m.Key,
			})
		}
	}
	return events, nil
}

func parseTempl(s string, m message) ([]byte, error) {
	t, err := template.New("").Parse(s)
	if err != nil {
		return []byte{}, err
	}

	var b bytes.Buffer
	if err := t.Execute(&b, m); err != nil {
		return []byte{}, err
	}

	return b.Bytes(), nil
}

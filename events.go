package main

import (
	"bytes"
	"regexp"
	"text/template"
)

func processMessage(m message, rules []rule, keyAliases map[string]string) ([]event, error) {
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
		if pass {
			for _, e := range r.Events {
				bEventType, err := parseTempl(e.EventType, m)
				if err != nil {
					return events, err
				}
				bKey, err := parseTempl(e.Key, m)
				if err != nil {
					return events, err
				}
				bKeyAlias, err := parseTempl(e.KeyAlias, m)
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

				key := string(bKey)
				if ka, ok := keyAliases[key]; len(ka) > 0 && ok {
					key = ka
				}
				if len(bKeyAlias) > 0 && len(bKey) > 0 {
					keyAliases[string(bKeyAlias)] = key
				}

				events = append(events, event{
					EventType: string(bEventType),
					Key:       key,
					SourceId:  string(bSourceId),
					TargetId:  string(bTargetId),
					Text:      string(bText),
					JSON:      m.Value,
				})
			}
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

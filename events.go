package main

import (
	"bytes"
	"regexp"
	"text/template"
)

func processMessage(m message, rules []rule, fsmIdAliases map[string]string, events *[]event, incompleteEvents *[]event, globalFSMId string) error {
	for _, r := range rules {
		pass := true
		for _, p := range r.Patterns {
			b, err := parseTempl(p.Field, m)
			if err != nil {
				return err
			}

			matched, err := regexp.Match(p.Pattern, b)
			if err != nil {
				return err
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
				return err
			}
			bFSMId, err := parseTempl(e.FSMId, m)
			if err != nil {
				return err
			}
			bFSMIdAlias, err := parseTempl(e.FSMIdAlias, m)
			if err != nil {
				return err
			}
			bSourceId, err := parseTempl(e.SourceId, m)
			if err != nil {
				return err
			}
			bTargetId, err := parseTempl(e.TargetId, m)
			if err != nil {
				return err
			}
			bText, err := parseTempl(e.Text, m)
			if err != nil {
				return err
			}

			fsmId := string(bFSMId)
			fsmIdAlias := string(bFSMIdAlias)
			if fa, ok := fsmIdAliases[fsmId]; len(fa) > 0 && ok {
				fsmId = fa
			}
			if _, ok := fsmIdAliases[fsmIdAlias]; !ok && len(fsmIdAlias) > 0 && len(fsmId) > 0 { // if new id/alias pair
				fsmIdAliases[fsmIdAlias] = fsmId // save new alias definition

				*events = append(*events, event{ // ui will need to resolve aliases too
					EventType:  "alias",
					FSMId:      fsmId,
					FSMIdAlias: fsmIdAlias,
				})

				for i, e := range *incompleteEvents { // fill in fsmIds on incomplete events
					if e.FSMIdAlias == fsmIdAlias {
						(*incompleteEvents)[i].FSMId = fsmId
						(*incompleteEvents)[i].FSMIdAlias = ""
					}
				}
			}

			json := []map[string]interface{}{}
			if !e.NoJSON {
				json = []map[string]interface{}{m.Value}
			}

			if len(fsmId) == 0 && len(fsmIdAlias) > 0 {
				*incompleteEvents = append(*incompleteEvents, event{
					EventType:  string(bEventType),
					FSMIdAlias: fsmIdAlias,
					SourceId:   string(bSourceId),
					TargetId:   string(bTargetId),
					Text:       string(bText),
					JSON:       json,
					Aggregate:  e.Aggregate,
				})
				continue
			}

			if len(globalFSMId) > 0 && globalFSMId != fsmId {
				continue
			}

			newE := event{
				EventType: string(bEventType),
				FSMId:     fsmId,
				SourceId:  string(bSourceId),
				TargetId:  string(bTargetId),
				Text:      string(bText),
				JSON:      json,
				Count:     1,
				Aggregate: e.Aggregate,
			}

			*events = aggregate(*events, newE, e.Aggregate, globalFSMId)
		}
	}
	return nil
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

func aggregate(events []event, e event, aggregate bool, globalFSMId string) []event {
	if len(globalFSMId) > 0 && globalFSMId != e.FSMId {
		return events
	}
	if !aggregate {
		return append(events, e)
	}

	for i, ev := range events {
		if ev.FSMId == e.FSMId && ev.SourceId == e.SourceId && ev.TargetId == e.TargetId {
			events[i].Count++
			events[i].JSON = append(events[i].JSON, e.JSON...)
			return events
		}
	}
	return append(events, e)
}

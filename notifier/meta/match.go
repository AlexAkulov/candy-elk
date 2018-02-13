package meta

import (
	"github.com/alexakulov/candy-elk"
)

const indexSuffixLength = len("-2016-02-10")

// MatchEvent matching decoded message for alert conditions
func (u *Updater) MatchEvent(m *elkstreams.DecodedLogMessage) []*AlertMeta {
	indexNameLength := len(m.IndexName)
	if indexNameLength < indexSuffixLength {
		return nil
	}
	indexTemplate := m.IndexName[0 : indexNameLength-indexSuffixLength]
	indexMetas, ok := u.alertMetas[indexTemplate]
	if !ok {
		return nil
	}

	ignored := false
	if excHashField, fieldOk := m.Fields["exc_stacktrace_hash"]; fieldOk {
		excHash := excHashField.(string)
		if metas, metasOk := u.eventMetas[indexTemplate]; metasOk {
			if meta, metaOk := metas[excHash]; metaOk {
				ignored = meta.Ignored
				if u.stats != nil && len(meta.Metric) > 0 {
					u.stats.Incr(meta.Metric, 1)
				}
			}
		}
	}

	var result []*AlertMeta
	for _, meta := range indexMetas {
		matched := 0
		if !meta.ApplyToIgnored && ignored {
			continue
		}
		for _, filter := range meta.Filters {
			field, ok := m.Fields[filter.Field]
			if !ok {
				break
			}
			if field == nil {
				u.Log.Debug("msg", "empty field", "field", filter.Field, "index", m.IndexName, "trigger", meta.Name)
				break
			}
			var (
				value float64
				bytes []byte
			)
			switch t := field.(type) {
			case string:
				bytes = []byte(field.(string))
			case []byte:
				bytes = field.([]byte)
			case int:
				value = float64(field.(int))
			case int32:
				value = float64(field.(int32))
			case int64:
				value = float64(field.(int64))
			case float32:
				value = float64(field.(float32))
			case float64:
				value = field.(float64)
			default:
				u.Log.Debug("msg", "Unsupported field type", "type", t, "index", m.IndexName, "trigger", meta.Name, "field", filter.Field)
				break
			}
			if filter.compiledRegexp != nil && !filter.compiledRegexp.Match(bytes) {
				break
			}
			if filter.compareFunction != nil {
				if len(bytes) > 0 {
					v, err := getFloatValue(filter.Field, string(bytes))
					if err != nil {
						u.Log.Debug("Can't get float64 value from %s of field %s", string(bytes), filter.Field)
						break
					}
					value = v
				}

				if !filter.compareFunction(value, filter.LimitValue) {
					break
				}
			}
			matched++
		}
		if matched == len(meta.Filters) {
			result = append(result, meta)
		}
	}
	return result
}

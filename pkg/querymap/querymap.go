package querymap

import (
	"github.com/mitchellh/mapstructure"
	"golang.org/x/exp/maps"
	"net/url"
	"slices"
	"strconv"
	"strings"
)

type anyList []any

// QueryMap - a map of query parameters. "any" can be one of: string, []string, QueryMap, anyList
type QueryMap map[string]any

func newQueryMap() QueryMap {
	return make(QueryMap)
}

func (q QueryMap) set(key string, untypedValue any) QueryMap {
	untypedEntry, ok := q[key]
	if !ok {
		q[key] = untypedValue
		return q
	}

	switch entry := untypedEntry.(type) {
	case string:
		switch value := untypedValue.(type) {
		case string: // string1 + string2 = []string{string1, string2}
			q[key] = []string{entry, value}
		case []string: // string1 + []string{string2, string3} = []string{string1, string2, string3}
			q[key] = append([]string{entry}, value...)
		case anyList: // string1 + []any{var1, var2} = []any{string1, var1, var2}
			q[key] = append(anyList{entry}, value...)
		case QueryMap: // string1 + {key1: val1, key2: val2} = []any{string1, {key1: val1, key2: val2}}
			q[key] = anyList{entry, value}
		}
	case []string:
		switch value := untypedValue.(type) {
		case string: // []string{string1, string2} + string3 = []string{string1, string2, string3}
			q[key] = append(entry, value)
		case []string: // []string{string1} + []string{string2} = []string{string1, string2}
			q[key] = append(entry, value...)
		case anyList: // []string{string1} + []any{var1, var2} = []any{string1, var1, var2}
			slc := anyList{}
			for _, s := range entry {
				slc = append(slc, s)
			}
			q[key] = append(slc, value...)
		case QueryMap: // []string{string1} + {key1: val1, key2: val2} = []any{string1, {key1: val1, key2: val2}}
			slc := anyList{}
			for _, s := range entry {
				slc = append(slc, s)
			}
			q[key] = append(slc, value)
		}
	case anyList:
		switch value := untypedValue.(type) {
		case string: // []any{var1, var2} + string3 = []any{var1, var2, string3}
			q[key] = append(entry, value)
		case []string: // []any{var1, var2} + []string{string3, string4} = []any{var1, var2, string3, string4}
			for _, s := range value {
				entry = append(entry, s)
			}
			q[key] = entry
		case anyList: // []any{var1, var2} + []any{var3, var4} = []any{var1, var2, var3, var4} or can merge(not needed)
			q[key] = append(entry, value...)
		case QueryMap: // []any{var1, var2} + {key1: val1} = []any{var1, var2, {key1: val1}} or can merge(not needed)
			q[key] = append(entry, value)
		}
	case QueryMap:
		switch value := untypedValue.(type) {
		case string: // {key1: val1} + string2 = []any{{key1: val1}, string2}
			q[key] = append(anyList{entry}, value)
		case []string: // {key1: val1} + []string{string2, string3} = []any{{key1: val1}, string2, string3}
			slc := anyList{entry}
			for _, s := range value {
				slc = append(slc, s)
			}
			q[key] = slc
		case anyList: // {key1: val1} + []any{var1, var2} = []any{{key1: val1}, var1, var2} or can merge(not needed)
			slc := anyList{entry}
			slc = append(slc, value...)
			q[key] = slc
		case QueryMap: // {key1: val1} + {key2: val2} = {key1: val1, key2: val2}
			for typedValueKey, typedValueValue := range value {
				entry.set(typedValueKey, typedValueValue)
			}
		}
	}

	return q
}

func nestedQuery(data QueryMap, key string, value []string) QueryMap {
	nextStart := strings.IndexRune(key, '[')
	nextEnd := strings.IndexRune(key, ']')

	currentKey := key

	if nextStart == -1 && nextEnd != -1 && len(key) == nextEnd+1 { // name]
		currentKey = key[:nextEnd]
	} else if nextStart != -1 && nextEnd != -1 && nextStart+1 == nextEnd { // name[]
		currentKey = key[:nextStart]
	} else if nextEnd+1 == nextStart { // name][
		currentKey = key[:nextEnd]
	} else if nextStart != -1 && nextStart < nextEnd { // name[a] or name[]
		currentKey = key[:nextStart]
	}

	// name[] or name][] and no any text after
	// regex: \[\]$ OR regex: \]\[$
	if nextStart+1 == nextEnd && nextEnd+1 == len(key) || nextEnd != -1 && nextStart > nextEnd && key[nextStart:] == "[]" && nextStart+2 == len(key) {
		return data.set(currentKey, value)
	}

	if nextStart != -1 {
		nextKey := ""
		if nextEnd != -1 && nextEnd < nextStart { // b][a
			nextKey = key[nextEnd+2:]
		} else { // b[]
			nextKey = key[nextStart+1:]
		}

		return data.set(currentKey, nestedQuery(newQueryMap(), nextKey, value))
	}
	if len(value) == 1 {
		return data.set(currentKey, value[0])
	}
	return data.set(currentKey, value)
}

func FromURL(URL *url.URL) QueryMap {
	data := newQueryMap()

	urlQuery := URL.Query()
	urlQueryKeys := maps.Keys(urlQuery)
	slices.Sort(urlQueryKeys)

	for _, key := range urlQueryKeys {
		value := urlQuery[key]

		nestedQuery(data, key, value)
	}

	for k, v := range data {
		data[k] = NormalizeSlicesNumbersIndexes(v)
	}

	return data
}

func ToStruct[T any](m QueryMap) (*T, error) {
	var result T

	config := &mapstructure.DecoderConfig{Metadata: nil, Result: &result, WeaklyTypedInput: true, TagName: "json"}
	decoder, _ := mapstructure.NewDecoder(config)
	if err := decoder.Decode(m); err != nil {
		return nil, err
	}

	return &result, nil
}

func FromURLToStruct[T any](URL *url.URL) (*T, error) {
	return ToStruct[T](FromURL(URL))
}

func FromURLStringToStruct[T any](URL string) (*T, error) {
	parsedUrl, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}

	return FromURLToStruct[T](parsedUrl)
}

func NormalizeSlicesNumbersIndexes(v any) any {
	switch value := v.(type) {
	case string:
		return value
	case []string:
		return value
	case QueryMap:
		total := 0
		keyAreNumbers := 0

		for key := range value {
			total++
			if _, err := strconv.Atoi(key); err == nil {
				keyAreNumbers++
			}
		}
		if total == keyAreNumbers {
			slc := anyList{}

			valueKeys := maps.Keys(value)
			slices.Sort(valueKeys)
			for _, valueKey := range valueKeys {
				slc = append(slc, value[valueKey])
			}

			return slc
		}

		for k, v := range value {
			value[k] = NormalizeSlicesNumbersIndexes(v)
		}

		return value
	case anyList:
		entry := make(anyList, len(value))
		for i, v := range value {
			entry[i] = NormalizeSlicesNumbersIndexes(v)
		}
		return entry
	}

	return v
}

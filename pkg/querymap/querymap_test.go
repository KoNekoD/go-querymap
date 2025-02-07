package querymap

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestFromURL(t *testing.T) {
	type args struct{ URL string }
	tests := []struct {
		name string
		args args
		want QueryMap
	}{
		{
			name: "simple",
			args: args{URL: "example.com?b=1&c=2"},
			want: QueryMap{"b": "1", "c": "2"},
		},
		{
			name: "array",
			args: args{URL: "example.com?b[]=1&b[]=2"},
			want: QueryMap{"b": []string{"1", "2"}},
		},
		{
			name: "nested array",
			args: args{URL: "example.com?b=1&b=2"},
			want: QueryMap{"b": []string{"1", "2"}},
		},
		{
			name: "nested",
			args: args{URL: "example.com?b[0]=1&b[1]=2"},
			want: QueryMap{"b": anyList{"1", "2"}},
		},
		{
			name: "empty",
			args: args{URL: "example.com"},
			want: QueryMap{},
		},
		{
			name: "empty value",
			args: args{URL: "example.com?b="},
			want: QueryMap{"b": ""},
		},
		{
			name: "empty value with nested",
			args: args{URL: "example.com?b[0]=&b[1]=2"},
			want: QueryMap{"b": anyList{"", "2"}},
		},
		{
			name: "nested without index",
			args: args{URL: "example.com?b[]=1&b[]=2"},
			want: QueryMap{"b": []string{"1", "2"}},
		},
		{
			name: "special characters",
			args: args{URL: "example.com?b=hello%20world"},
			want: QueryMap{"b": "hello world"},
		},
		{
			name: "multiple nested keys",
			args: args{URL: "example.com?a[b][c]=1&a[b][d]=2"},
			want: QueryMap{
				"a": QueryMap{
					"b": QueryMap{
						"c": "1",
						"d": "2",
					},
				},
			},
		},
		{
			name: "array with object",
			args: args{URL: "example.com?b[0][c]=1&b[0][d]=2"},
			want: QueryMap{
				"b": anyList{
					QueryMap{
						"c": "1",
						"d": "2",
					},
				},
			},
		},
		{
			name: "duplicate keys",
			args: args{URL: "example.com?b=1&b=2"},
			want: QueryMap{"b": []string{"1", "2"}},
		},
		{
			name: "complex query",
			args: args{URL: "example.com?a[b]=1&a[c][d][e]=2"},
			want: QueryMap{
				"a": QueryMap{
					"b": "1",
					"c": QueryMap{
						"d": QueryMap{
							"e": "2",
						},
					},
				},
			},
		},
		{
			name: "numeric keys",
			args: args{URL: "example.com?123=hello"},
			want: QueryMap{"123": "hello"},
		},
		{
			name: "boolean values",
			args: args{URL: "example.com?flag=true&enabled=false"},
			want: QueryMap{"flag": "true", "enabled": "false"},
		},
		{
			name: "encoded brackets",
			args: args{URL: "example.com?b%5B0%5D=1&b%5B1%5D=2"},
			want: QueryMap{"b": anyList{"1", "2"}},
		},
		{
			name: "deep nesting",
			args: args{URL: "example.com?a[b][c][d][e]=value"},
			want: QueryMap{
				"a": QueryMap{
					"b": QueryMap{
						"c": QueryMap{
							"d": QueryMap{
								"e": "value",
							},
						},
					},
				},
			},
		},
		{
			name: "empty array",
			args: args{URL: "example.com?b[]="},
			want: QueryMap{"b": []string{""}},
		},
		{
			name: "special characters in key",
			args: args{URL: "example.com?key%21=value"},
			want: QueryMap{"key!": "value"},
		},
		{
			name: "encoded characters in value",
			args: args{URL: "example.com?key=hello%2C%20world%21"},
			want: QueryMap{"key": "hello, world!"},
		},
		{
			name: "empty key",
			args: args{URL: "example.com?=value"},
			want: QueryMap{"": "value"},
		},
		{
			name: "nested arrays",
			args: args{URL: "example.com?b[0][]=1&b[0][]=2&b[1][]=3"},
			want: QueryMap{
				"b": anyList{
					[]string{
						"1",
						"2",
					},
					[]string{
						"3",
					},
				},
			},
		},
		{
			name: "non-standard format",
			args: args{URL: "example.com?b=1&&c=2"},
			want: QueryMap{"b": "1", "c": "2"},
		},
		{
			name: "repeated nested keys",
			args: args{URL: "example.com?a[b]=1&a[b]=2"},
			want: QueryMap{
				"a": QueryMap{
					"b": []string{
						"1",
						"2",
					},
				},
			},
		},
		{
			name: "hydrate object",
			args: args{URL: "example.com?pagination[query][orders]=1&pagination[query]=1&pagination=1&pagination=2"},
			want: QueryMap{
				"pagination": anyList{
					"1",
					"2",
					QueryMap{
						"query": "1",
					},
					QueryMap{
						"query": QueryMap{
							"orders": "1",
						},
					},
				},
			},
		},
		{
			name: "add string to string slice",
			args: args{URL: "example.com?b[]=1&b=2"},
			want: QueryMap{"b": []string{"2", "1"}},
		},
		{
			name: "add string slice to string",
			args: args{URL: "example.com?b=1&b[]=2"},
			want: QueryMap{"b": []string{"1", "2"}},
		},
		{
			name: "add string slice to string slice",
			args: args{URL: "example.com?b=1&b=2&b[]=3"},
			want: QueryMap{"b": []string{"1", "2", "3"}},
		},
		{
			name: "add nested string slice to nested string slice",
			args: args{URL: "example.com?data[b]=1&data[b]=2&data[b][]=3"},
			want: QueryMap{"data": QueryMap{"b": []string{"1", "2", "3"}}},
		},
		{
			name: "nested struct",
			args: args{URL: "example.com?filter[name]=Ken&pagination[startFrom]=984&pagination[limit]=25"},
			want: QueryMap{
				"filter": QueryMap{
					"name": "Ken",
				},
				"pagination": QueryMap{
					"startFrom": "984",
					"limit":     "25",
				},
			},
		},
		{
			name: "many nested structs",
			args: args{URL: "example.com?filter[name]=Ken&pagination[startFrom]=984&pagination[limit]=25&filter[name]=Ken2&pagination[startFrom]=982&pagination[limit]=256"},
			want: QueryMap{
				"filter": QueryMap{
					"name": []string{"Ken", "Ken2"},
				},
				"pagination": QueryMap{
					"startFrom": []string{"984", "982"},
					"limit":     []string{"25", "256"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				parsedUrl, err := url.Parse(tt.args.URL)
				if err != nil {
					t.Fatal(err)
				}
				if got := FromURL(parsedUrl); !reflect.DeepEqual(got, tt.want) {
					q := parsedUrl.Query()
					gotJson, _ := json.Marshal(got)
					wantJson, _ := json.Marshal(tt.want)
					t.Errorf(
						"FromURL() = %v, want %v, \n\nG\t %v \nW\t %v \n\n\t %v",
						got,
						tt.want,
						string(gotJson),
						string(wantJson),
						q,
					)
				}
			},
		)
	}
}

func TestQueryMapSet(t *testing.T) {
	const sharedKey = "a"

	var (
		emptyQm   = func() QueryMap { return QueryMap{} }
		stringQm  = func() QueryMap { return QueryMap{sharedKey: "b1"} }
		stringsQm = func() QueryMap { return QueryMap{sharedKey: []string{"b1", "b2"}} }
		anyListQm = func() QueryMap { return QueryMap{sharedKey: anyList{"b1", "b2"}} }
		mapQm     = func() QueryMap { return QueryMap{sharedKey: QueryMap{"b": "1"}} }
		stringIn  = func() string { return "v2" }
		stringsIn = func() []string { return []string{"v2", "v3"} }
		anyListIn = func() anyList { return anyList{"v2", "v3"} }
		mapIn     = func() QueryMap { return QueryMap{"v": "1"} }
	)

	type args struct {
		qm    QueryMap
		key   string
		value interface{}
	}
	tests := []struct {
		name string
		args args
		want QueryMap
	}{
		{
			name: "set to not set key",
			args: args{qm: emptyQm(), key: sharedKey, value: stringIn()},
			want: QueryMap{sharedKey: "v2"},
		},
		{
			name: "set to string value of type string",
			args: args{qm: stringQm(), key: sharedKey, value: stringIn()},
			want: QueryMap{sharedKey: []string{"b1", "v2"}},
		},
		{
			name: "set to string value of type []string",
			args: args{qm: stringQm(), key: sharedKey, value: stringsIn()},
			want: QueryMap{sharedKey: []string{"b1", "v2", "v3"}},
		},
		{
			name: "set to string value of type anyList",
			args: args{qm: stringQm(), key: sharedKey, value: anyListIn()},
			want: QueryMap{sharedKey: anyList{"b1", "v2", "v3"}},
		},
		{
			name: "set to string value of type QueryMap",
			args: args{qm: stringQm(), key: sharedKey, value: mapIn()},
			want: QueryMap{sharedKey: anyList{"b1", QueryMap{"v": "1"}}},
		},
		{
			name: "set to []string value of type string",
			args: args{qm: stringsQm(), key: sharedKey, value: stringIn()},
			want: QueryMap{sharedKey: []string{"b1", "b2", "v2"}},
		},
		{
			name: "set to []string value of type []string",
			args: args{qm: stringsQm(), key: sharedKey, value: stringsIn()},
			want: QueryMap{sharedKey: []string{"b1", "b2", "v2", "v3"}},
		},
		{
			name: "set to []string value of type anyList",
			args: args{qm: stringsQm(), key: sharedKey, value: anyListIn()},
			want: QueryMap{sharedKey: anyList{"b1", "b2", "v2", "v3"}},
		},
		{
			name: "set to []string value of type QueryMap",
			args: args{qm: stringsQm(), key: sharedKey, value: mapIn()},
			want: QueryMap{sharedKey: anyList{"b1", "b2", QueryMap{"v": "1"}}},
		},
		{
			name: "set to anyList value of type string",
			args: args{qm: anyListQm(), key: sharedKey, value: stringIn()},
			want: QueryMap{sharedKey: anyList{"b1", "b2", "v2"}},
		},
		{
			name: "set to anyList value of type []string",
			args: args{qm: anyListQm(), key: sharedKey, value: stringsIn()},
			want: QueryMap{sharedKey: anyList{"b1", "b2", "v2", "v3"}},
		},
		{
			name: "set to anyList value of type anyList",
			args: args{qm: anyListQm(), key: sharedKey, value: anyListIn()},
			want: QueryMap{sharedKey: anyList{"b1", "b2", "v2", "v3"}},
		},
		{
			name: "set to anyList value of type QueryMap",
			args: args{qm: anyListQm(), key: sharedKey, value: mapIn()},
			want: QueryMap{sharedKey: anyList{"b1", "b2", QueryMap{"v": "1"}}},
		},
		{
			name: "set to QueryMap value of type string",
			args: args{qm: mapQm(), key: sharedKey, value: stringIn()},
			want: QueryMap{sharedKey: anyList{QueryMap{"b": "1"}, "v2"}},
		},
		{
			name: "set to QueryMap value of type []string",
			args: args{qm: mapQm(), key: sharedKey, value: stringsIn()},
			want: QueryMap{sharedKey: anyList{QueryMap{"b": "1"}, "v2", "v3"}},
		},
		{
			name: "set to QueryMap value of type anyList",
			args: args{qm: mapQm(), key: sharedKey, value: anyListIn()},
			want: QueryMap{sharedKey: anyList{QueryMap{"b": "1"}, "v2", "v3"}},
		},
		{
			name: "set to QueryMap value of type QueryMap",
			args: args{qm: mapQm(), key: sharedKey, value: mapIn()},
			want: QueryMap{sharedKey: QueryMap{"b": "1", "v": "1"}},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := tt.args.qm.set(tt.args.key, tt.args.value); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("set() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestNormalizeSlicesNumbersIndexes(t *testing.T) {
	qm := QueryMap{
		"0": "1",
		"2": "3",
		"4": "5",
	}
	want := anyList{"1", "3", "5"}
	if got := NormalizeSlicesNumbersIndexes(qm); !reflect.DeepEqual(got, want) {
		t.Errorf("NormalizeSlicesNumbersIndexes() = %v, want %v", got, want)
	}

	if NormalizeSlicesNumbersIndexes(nil) != nil {
		t.Errorf("NormalizeSlicesNumbersIndexes() = %v, want %v", nil, nil)
	}
}

func TestBenchmarkFromURL(t *testing.T) {
	t.Run(
		"benchmark default", func(t *testing.T) {
			rawUrl := strings.Builder{}
			rawUrl.WriteString("https://example.com?nestedData=1")
			URL, err := url.Parse(rawUrl.String())
			if err != nil {
				t.Fatal(err)
			}
			b := testing.Benchmark(
				func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						_ = FromURL(URL)
					}
				},
			)
			t.Logf("BenchmarkFromURL: %v", b)
		},
	)

	t.Run(
		"benchmark nested", func(t *testing.T) {
			rawUrl := strings.Builder{}
			rawUrl.WriteString("https://example.com")
			rawUrl.WriteString("?")
			rawUrl.WriteString("nestedData")
			for i := 0; i < 100; i++ {
				rawUrl.WriteString(fmt.Sprintf("[%d]", i))
			}
			rawUrl.WriteString("=")
			rawUrl.WriteString("HelloWorld")
			URL, err := url.Parse(rawUrl.String())
			if err != nil {
				t.Fatal(err)
			}
			b := testing.Benchmark(
				func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						_ = FromURL(URL)
					}
				},
			)
			t.Logf("BenchmarkFromURL: %v", b)
		},
	)
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

type TestReadMeT1 struct {
	Names []string `json:"names,omitempty"`
}

type TestReadMeT2 struct {
	Names []struct {
		FirstName string `json:"firstName,omitempty"`
	} `json:"names,omitempty"`
}

func TestReadMe(t *testing.T) {
	str1 := "site.com?names[0]=John"

	str2 := "site.com?names[]=John"

	str3 := "site.com?names[0][firstName]=John"

	str4 := "site.com?names[][firstName]=John"

	dto1, err := FromURLStringToStruct[TestReadMeT1](str1)
	panicIfErr(err)

	dto2, err := FromURLStringToStruct[TestReadMeT1](str2)
	panicIfErr(err)

	dto3, err := FromURLStringToStruct[TestReadMeT2](str3)
	panicIfErr(err)

	parsedUrl, err := url.Parse(str4)
	panicIfErr(err)
	qm4 := FromURL(parsedUrl)
	dto4, err := ToStruct[TestReadMeT2](qm4)
	panicIfErr(err)

	exceptedDto1 := &TestReadMeT1{Names: []string{"John"}}
	exceptedDto2 := &TestReadMeT1{Names: []string{"John"}}
	exceptedDto3 := &TestReadMeT2{
		Names: []struct {
			FirstName string `json:"firstName,omitempty"`
		}{{FirstName: "John"}},
	}
	exceptedDto4 := &TestReadMeT2{
		Names: []struct {
			FirstName string `json:"firstName,omitempty"`
		}{{FirstName: "John"}},
	}

	if !reflect.DeepEqual(dto1, exceptedDto1) {
		t.Errorf("Expected dto1 to be %v, got %v", exceptedDto1, dto1)
	}
	if !reflect.DeepEqual(dto2, exceptedDto2) {
		t.Errorf("Expected dto2 to be %v, got %v", exceptedDto2, dto2)
	}
	if !reflect.DeepEqual(dto3, exceptedDto3) {
		t.Errorf("Expected dto3 to be %v, got %v", exceptedDto3, dto3)
	}
	if !reflect.DeepEqual(dto4, exceptedDto4) { // Is not possible to decode `names[][firstName]=John`
		/**
		 * Technically it is possible to realize its decoding, but it is intentionally not done to avoid abuse of
		 *  abbreviations, such practice has a nuance associated with the fact that it is impossible to understand
		 *  under what index should go an element of the nested string.
		 *
		 * Suppose we have an array of dtos with many fields, we need to know under which index to fill
		 *  the field of the nested dtos of the tree we are parsing. Without an index it is impossible,
		 *  so we can omit indexes only at the end of the string.
		 *
		 * Technically it is possible to realize parsing of a string on ordered map in and pre-calculate indexes,
		 *  but it can have an effect on performance because of more complex parsing, so here we go :)
		 */
		exceptedQm4 := QueryMap{
			"names": QueryMap{
				"": QueryMap{
					"firstName": "John",
				},
			},
		}
		if !reflect.DeepEqual(qm4, exceptedQm4) {
			t.Errorf("Expected qm4 to be %v, got %v", exceptedQm4, qm4)
		}
	}
}

func TestFromURLStringToStructParseError(t *testing.T) {
	_, err := FromURLStringToStruct[any](":")
	if err == nil {
		t.Errorf("Expected error, got nil")
	} else {
		if err.Error() != "parse \":\": missing protocol scheme" {
			t.Errorf("Expected error to be 'missing protocol scheme', got %v", err)
		}
	}
}

func TestToStructError(t *testing.T) {
	_, err := ToStruct[struct {
		Name complex64 `json:"name"`
	}](QueryMap{"name": 1})
	if err == nil {
		t.Errorf("Expected error, got nil")
	} else {
		const exceptedErr = "1 error(s) decoding:\n\n* name: unsupported type: complex64"
		if err.Error() != exceptedErr {
			t.Errorf("Expected error to be '%s', got %v", exceptedErr, err)
		}
	}
}

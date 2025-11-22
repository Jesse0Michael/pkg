package test

import (
	"fmt"
	"reflect"
	"testing"
)

type Person struct {
	Name string
	Age  int
}

func ExampleLoadFile_variable() {
	t := &testing.T{}

	me := LoadFile(t, "testdata/person.json")
	fmt.Println(string(me))
	// Output:
	// {
	//     "name": "Jesse Michael",
	//     "age": 29
	// }
}

func ExampleLoadJSONFile_variable() {
	t := &testing.T{}

	me := LoadJSONFile[Person](t, "testdata/person.json")
	fmt.Println(me)
	// Output:
	// {Jesse Michael 29}
}

func ExampleLoadJSONFile_casting() {
	t := &testing.T{}

	fmt.Println(*LoadJSONFile[*Person](t, "testdata/person.json"))
	// Output:
	// {Jesse Michael 29}
}

func ExampleLoadTemplate() {
	t := &testing.T{}

	data := struct {
		Name string
		Age  int
	}{
		Name: "Jesse Michael",
		Age:  29,
	}

	fmt.Println(string(LoadTemplate(t, "testdata/person.json.tmpl", data)))
	// Output:
	// {
	//     "name": "Jesse Michael",
	//     "age": 29
	// }
}

func ExampleLoadJSONTemplate() {
	t := &testing.T{}

	data := struct {
		Name string
		Age  int
	}{
		Name: "Jesse Michael",
		Age:  29,
	}

	fmt.Println(LoadJSONTemplate[Person](t, "testdata/person.json.tmpl", data))
	// Output:
	// {Jesse Michael 29}
}

func TestLoadJSONFile(t *testing.T) {
	expected := Person{
		Name: "Jesse Michael",
		Age:  29,
	}

	actual := LoadJSONFile[Person](t, "testdata/person.json")
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("LoadJSONFile[Person]() = %v, want %v", actual, expected)
	}

	ptr := LoadJSONFile[*Person](t, "testdata/person.json")
	if !reflect.DeepEqual(ptr, &expected) {
		t.Errorf("LoadJSONFile[*Person]() = %v, want %v", ptr, &expected)
	}
}

func TestLoadFile_FailToLoad(t *testing.T) {
	tt := &testing.T{}
	LoadFile(tt, "testdata/notfound.json")
	if !tt.Failed() {
		t.Error("expected LoadJSONFile() to fail to load file")
	}
}

func TestLoadJSONFile_FailToUnmarshal(t *testing.T) {
	tt := &testing.T{}
	LoadJSONFile[Person](tt, "testdata/not.json")
	if !tt.Failed() {
		t.Error("expected LoadJSONFile() to fail to unmarshal json")
	}
}

func TestLoadTemplate_FailToParse(t *testing.T) {
	tt := &testing.T{}
	LoadTemplate(tt, "testdata/invalid.tmpl", struct{ Name string }{Name: "bad"})
	if !tt.Failed() {
		t.Error("expected LoadTemplate() to fail to parse template")
	}
}

func TestLoadJSONTemplate_FailToUnmarshal(t *testing.T) {
	tt := &testing.T{}
	LoadJSONTemplate[Person](tt, "testdata/not.json", nil)
	if !tt.Failed() {
		t.Error("expected LoadJSONTemplate() to fail to unmarshal json")
	}
}

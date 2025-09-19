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

	var me Person
	LoadJSONFile(t, "testdata/person.json", &me)
	fmt.Println(me)
	// Output:
	// {Jesse Michael 29}
}

func ExampleLoadJSONFile_casting() {
	t := &testing.T{}

	fmt.Println(*LoadJSONFile(t, "testdata/person.json", &Person{}).(*Person))
	// Output:
	// {Jesse Michael 29}
}

func TestLoadJSONFile(t *testing.T) {
	expected := Person{
		Name: "Jesse Michael",
		Age:  29,
	}

	var actual Person
	r := LoadJSONFile(t, "testdata/person.json", &actual).(*Person)

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("LoadJSONFile() parameter = %v, want %v", actual, expected)
	}
	if !reflect.DeepEqual(r, &expected) {
		t.Errorf("LoadJSONFile() result = %v, want %v", r, &expected)
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
	LoadJSONFile(tt, "testdata/not.json", &Person{})
	if !tt.Failed() {
		t.Error("expected LoadJSONFile() to fail to unmarshal json")
	}
}

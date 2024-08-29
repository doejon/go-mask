package mask

import (
	"fmt"
	"reflect"
	"testing"
)

func ExampleMust() {
	tests := []interface{}{
		`"Now cut that out!"`,
		39,
		true,
		false,
		2.14,
		[]string{
			"Phil Harris",
			"Rochester van Jones",
			"Mary Livingstone",
			"Dennis Day",
		},
		[2]string{
			"Jell-O",
			"Grape-Nuts",
		},
	}

	for _, expected := range tests {
		actual := Must(expected)
		fmt.Println(actual)
	}
	// Output:
	// "Now cut that out!"
	// 39
	// true
	// false
	// 2.14
	// [Phil Harris Rochester van Jones Mary Livingstone Dennis Day]
	// [Jell-O Grape-Nuts]
}

type Foo struct {
	Foo *Foo
	Bar int
}

func ExampleMap() {
	x := map[string]*Foo{
		"foo": {Bar: 1},
		"bar": {Bar: 2},
	}
	y := Must(x)
	for _, k := range []string{"foo", "bar"} { // to ensure consistent order
		fmt.Printf("x[\"%v\"] = y[\"%v\"]: %v\n", k, k, x[k] == y[k])
		fmt.Printf("x[\"%v\"].Foo = y[\"%v\"].Foo: %v\n", k, k, x[k].Foo == y[k].Foo)
		fmt.Printf("x[\"%v\"].Bar = y[\"%v\"].Bar: %v\n", k, k, x[k].Bar == y[k].Bar)
	}
	// Output:
	// x["foo"] = y["foo"]: false
	// x["foo"].Foo = y["foo"].Foo: false
	// x["foo"].Bar = y["foo"].Bar: true
	// x["bar"] = y["bar"]: false
	// x["bar"].Foo = y["bar"].Foo: false
	// x["bar"].Bar = y["bar"].Bar: true
}

func TestInterface(t *testing.T) {
	x := []interface{}{nil}
	y := Must(x)
	if !reflect.DeepEqual(x, y) || len(y) != 1 {
		t.Errorf("expect %v == %v; y had length %v (expected 1)", x, y, len(y))
	}
	var a interface{}
	b := Must(a)
	if a != b {
		t.Errorf("expected %v == %v", a, b)
	}
}

func ExampleAvoidInfiniteLoops() {
	x := &Foo{
		Bar: 4,
	}
	x.Foo = x
	y := Must(x)
	fmt.Printf("x == y: %v\n", x == y)
	fmt.Printf("x == x.Foo: %v\n", x == x.Foo)
	fmt.Printf("y == y.Foo: %v\n", y == y.Foo)
	// Output:
	// x == y: false
	// x == x.Foo: true
	// y == y.Foo: true
}

func TestUnsupportedKind(t *testing.T) {
	x := func() {}

	tests := []interface{}{
		x,
		map[bool]interface{}{true: x},
		[]interface{}{x},
	}

	for _, test := range tests {
		y, err := Mask(test)
		if y != nil {
			t.Errorf("expected %v to be nil", y)
		}
		if err == nil {
			t.Errorf("expected err to not be nil")
		}
	}
}

func TestUnsupportedKindPanicsOnMust(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected a panic; didn't get one")
		}
	}()
	x := func() {}
	Must(x)
}

func TestMismatchedTypesFail(t *testing.T) {
	tests := []struct {
		input interface{}
		kind  reflect.Kind
	}{
		{
			map[int]int{1: 2, 2: 4, 3: 8},
			reflect.Map,
		},
		{
			[]int{2, 8},
			reflect.Slice,
		},
	}
	for _, test := range tests {
		for kind, copier := range copiers {
			if kind == test.kind {
				continue
			}
			actual, err := copier(test.input, nil)
			if actual != nil {

				t.Errorf("%v attempted value %v as %v; should be nil value, got %v", test.kind, test.input, kind, actual)
			}
			if err == nil {
				t.Errorf("%v attempted value %v as %v; should have gotten an error", test.kind, test.input, kind)
			}
		}
	}
}

func TestTwoNils(t *testing.T) {
	type Foo struct {
		A int
	}
	type Bar struct {
		B int
	}
	type FooBar struct {
		Foo  *Foo
		Bar  *Bar
		Foo2 *Foo
		Bar2 *Bar
	}

	src := &FooBar{
		Foo2: &Foo{1},
		Bar2: &Bar{2},
	}

	dst := Must(src)

	if !reflect.DeepEqual(src, dst) {
		t.Errorf("expect %v == %v; ", src, dst)
	}

}

type TestString string

func (t TestString) MaskXXX() TestString {
	return TestString("MASKED")
}

type testInt int

func (t testInt) MaskXXX() testInt {
	return 0
}

type testMap map[string]string

func (t testMap) MaskXXX() testMap {
	return map[string]string{}
}

type testSlice []string

func (t testSlice) MaskXXX() testSlice {
	return nil
}

type testStruct struct {
	S1 TestString
	S2 *TestString

	I1 testInt
	I2 *testInt
	I3 *testInt
	Mp testMap
	Sl testSlice

	Value string

	Strct1 testStruct2
	Strct2 *testStruct2

	CustomInterface interface{}
}

func newTestStruct() *testStruct {
	var ts2 TestString = "test string 2"
	var ti2 testInt = 2
	var _ = ts2
	var _ = ti2

	return &testStruct{
		S1:    "test string",
		S2:    &ts2,
		I1:    1,
		I2:    &ti2,
		I3:    nil,
		Mp:    map[string]string{"testKey": "testValue"},
		Sl:    []string{"sensitive"},
		Value: "test value",
		Strct1: testStruct2{
			N: "n1",
		},
		Strct2: &testStruct2{
			N: "n2",
		},
	}
}

func (t *testStruct) MaskXXX() {
	t.Value = "MASKED"
}

type testStruct2 struct {
	N string
}

func (t testStruct2) MaskXXX() testStruct2 {
	return testStruct2{
		N: "MASKED",
	}
}

func TestMask(t *testing.T) {

	val := newTestStruct()
	val2 := newTestStruct()
	masked := Must(val)

	// make sure the original value stays untouched
	if !reflect.DeepEqual(val, val2) {
		t.Errorf("expect %v == %v; ", val, val2)
	}

	if reflect.DeepEqual(val, masked) {
		t.Errorf("expect %v != %v", val, masked)
	}

	if masked.S1 != "MASKED" {
		t.Errorf("expect %v == MASKED", masked.S1)
	}
	if *masked.S2 != "MASKED" {
		t.Errorf("expect %v == MASKED", masked.S2)
	}
	if masked.I1 != 0 {
		t.Errorf("expect %v == 0", masked.I1)
	}
	if len(masked.Mp) != 0 {
		t.Errorf("expect %v == 0", masked.Mp)
	}

	if len(masked.Sl) != 0 {
		t.Errorf("expect %v == 0", masked.Sl)
	}

	if *masked.I2 != 0 {
		t.Errorf("expect %v == 0", *masked.I2)
	}

	if masked.I3 != nil {
		t.Errorf("expect %v == nil", *masked.I3)
	}

	if masked.Strct1.N != "MASKED" {
		t.Errorf("expect %v == MASKED", masked.Strct1.N)
	}

	if masked.Strct2.N != "MASKED" {
		t.Errorf("expect %v == MASKED", masked.Strct2.N)
	}

	if masked.Value != "MASKED" {
		t.Errorf("expect %v == MASKED", masked.Value)
	}

}

func TestStructInterfaceKey(t *testing.T) {

	val := newTestStruct()
	val.CustomInterface = "123"
	masked := Must(val)

	if masked.CustomInterface != "123" {
		t.Errorf("expect %v == 123", masked.CustomInterface)
	}

	val.CustomInterface = 1
	masked = Must(val)

	if masked.CustomInterface != 1 {
		t.Errorf("expect %v == 1", masked.CustomInterface)
	}

	val.CustomInterface = []string{"1"}
	masked = Must(val)

	arr, ok := masked.CustomInterface.([]string)
	if !ok {
		t.Errorf("expect %v to be array", masked.CustomInterface)
	}
	if len(arr) != 1 {
		t.Errorf("expect %v to have len 1", masked.CustomInterface)
	}
	if arr[0] != "1" {
		t.Errorf("expect %v == 1", arr[0])
	}

	type S struct {
		Data string
	}
	var x = &S{
		Data: "123",
	}

	val.CustomInterface = x
	masked = Must(val)

	ptr, ok := masked.CustomInterface.(*S)
	if !ok {
		t.Fatalf("expect %v to be ptr S", masked.CustomInterface)
	}
	if ptr == nil {
		t.Fatalf("expected ptr to be copied, got nil")
	}

	if ptr.Data != "123" {
		if arr[0] != "1" {
			t.Errorf("expect %v == 123", ptr.Data)
		}
	}

	val.CustomInterface = *x
	masked = Must(val)

	strct, ok := masked.CustomInterface.(S)
	if !ok {
		t.Fatalf("expect %v to be  S", masked.CustomInterface)
	}

	if strct.Data != "123" {
		if arr[0] != "1" {
			t.Errorf("expect %v == 123", strct.Data)
		}
	}

}

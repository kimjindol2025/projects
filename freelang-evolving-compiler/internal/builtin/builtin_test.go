package builtin

import (
	"testing"
)

// TestLookupBuiltin tests basic builtin lookup
func TestLookupBuiltin(t *testing.T) {
	def, ok := Lookup("print")
	if !ok {
		t.Errorf("print builtin not found")
	}
	if def.Name != "print" {
		t.Errorf("expected name 'print', got %q", def.Name)
	}
	if def.ReturnTypeName != "unit" {
		t.Errorf("expected unit type, got %q", def.ReturnTypeName)
	}
}

// TestIsBuiltin tests IsBuiltin check
func TestIsBuiltin(t *testing.T) {
	if !IsBuiltin("print") {
		t.Errorf("print should be a builtin")
	}
	if IsBuiltin("nonexistent") {
		t.Errorf("nonexistent should not be a builtin")
	}
}

// TestPrintBuiltin tests print() function
func TestPrintBuiltin(t *testing.T) {
	def, _ := Lookup("print")
	result := def.Impl("hello")
	if result != nil {
		t.Errorf("print should return nil (unit), got %v", result)
	}
}

// TestPrintlnBuiltin tests println() function
func TestPrintlnBuiltin(t *testing.T) {
	def, _ := Lookup("println")
	result := def.Impl("world")
	if result != nil {
		t.Errorf("println should return nil (unit), got %v", result)
	}
}

// TestStringLength tests len_str() function
func TestStringLength(t *testing.T) {
	def, _ := Lookup("len_str")
	tests := []struct {
		input    string
		expected int
	}{
		{"", 0},
		{"hello", 5},
		{"안녕하세요", 5},
	}
	for _, test := range tests {
		result := def.Impl(test.input)
		if result != test.expected {
			t.Errorf("len_str(%q) = %v, expected %d", test.input, result, test.expected)
		}
	}
}

// TestStringConcat tests concat() function
func TestStringConcat(t *testing.T) {
	def, _ := Lookup("concat")
	tests := []struct {
		a, b     string
		expected string
	}{
		{"", "", ""},
		{"hello", "world", "helloworld"},
		{"a", "b", "ab"},
	}
	for _, test := range tests {
		result := def.Impl(test.a, test.b)
		if result != test.expected {
			t.Errorf("concat(%q, %q) = %v, expected %q", test.a, test.b, result, test.expected)
		}
	}
}

// TestStringSubstring tests substring() function
func TestStringSubstring(t *testing.T) {
	def, _ := Lookup("substring")
	tests := []struct {
		s        string
		start    int
		end      int
		expected string
	}{
		{"hello", 0, 2, "he"},
		{"hello", 1, 4, "ell"},
		{"hello", 0, 5, "hello"},
	}
	for _, test := range tests {
		result := def.Impl(test.s, test.start, test.end)
		if result != test.expected {
			t.Errorf("substring(%q, %d, %d) = %v, expected %q", test.s, test.start, test.end, result, test.expected)
		}
	}
}

// TestStringUpper tests upper() function
func TestStringUpper(t *testing.T) {
	def, _ := Lookup("upper")
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "HELLO"},
		{"HELLO", "HELLO"},
		{"HeLLo", "HELLO"},
	}
	for _, test := range tests {
		result := def.Impl(test.input)
		if result != test.expected {
			t.Errorf("upper(%q) = %v, expected %q", test.input, result, test.expected)
		}
	}
}

// TestStringLower tests lower() function
func TestStringLower(t *testing.T) {
	def, _ := Lookup("lower")
	tests := []struct {
		input    string
		expected string
	}{
		{"HELLO", "hello"},
		{"hello", "hello"},
		{"HeLLo", "hello"},
	}
	for _, test := range tests {
		result := def.Impl(test.input)
		if result != test.expected {
			t.Errorf("lower(%q) = %v, expected %q", test.input, result, test.expected)
		}
	}
}

// TestStringSplit tests split() function
func TestStringSplit(t *testing.T) {
	def, _ := Lookup("split")
	result := def.Impl("a,b,c", ",")
	parts, ok := result.([]string)
	if !ok {
		t.Errorf("split should return []string, got %T", result)
	}
	if len(parts) != 3 {
		t.Errorf("split should return 3 parts, got %d", len(parts))
	}
	if parts[0] != "a" || parts[1] != "b" || parts[2] != "c" {
		t.Errorf("split parts incorrect: %v", parts)
	}
}

// TestArrayLength tests len_arr() function
func TestArrayLength(t *testing.T) {
	def, _ := Lookup("len_arr")
	arr := []interface{}{1, 2, 3}
	result := def.Impl(arr)
	if result != 3 {
		t.Errorf("len_arr([1,2,3]) = %v, expected 3", result)
	}
}

// TestArrayAppend tests append() function
func TestArrayAppend(t *testing.T) {
	def, _ := Lookup("append")
	arr := []interface{}{1, 2}
	result := def.Impl(arr, 3)
	newArr, ok := result.([]interface{})
	if !ok {
		t.Errorf("append should return []interface{}, got %T", result)
	}
	if len(newArr) != 3 {
		t.Errorf("append should return 3-element array, got %d", len(newArr))
	}
	if newArr[2] != 3 {
		t.Errorf("append should add element at end, got %v", newArr[2])
	}
}

// TestArrayGet tests get() function
func TestArrayGet(t *testing.T) {
	def, _ := Lookup("get")
	arr := []interface{}{10, 20, 30}
	result := def.Impl(arr, 1)
	if result != 20 {
		t.Errorf("get(arr, 1) = %v, expected 20", result)
	}
}

// TestArraySet tests set() function
func TestArraySet(t *testing.T) {
	def, _ := Lookup("set")
	arr := []interface{}{10, 20, 30}
	def.Impl(arr, 1, 99)
	if arr[1] != 99 {
		t.Errorf("set should modify array element, got %v", arr[1])
	}
}

// TestArraySlice tests slice() function
func TestArraySlice(t *testing.T) {
	def, _ := Lookup("slice")
	arr := []interface{}{1, 2, 3, 4, 5}
	result := def.Impl(arr, 1, 4)
	sliced, ok := result.([]interface{})
	if !ok {
		t.Errorf("slice should return []interface{}, got %T", result)
	}
	if len(sliced) != 3 {
		t.Errorf("slice(arr, 1, 4) should return 3-element array, got %d", len(sliced))
	}
	if sliced[0] != 2 || sliced[1] != 3 || sliced[2] != 4 {
		t.Errorf("slice result incorrect: %v", sliced)
	}
}

// TestAllDefs tests that all builtins are registered
func TestAllDefs(t *testing.T) {
	defs := AllDefs()
	expectedCount := 13 // print, println, 6 string, 5 array
	if len(defs) != expectedCount {
		t.Errorf("expected %d builtins, got %d", expectedCount, len(defs))
	}
}

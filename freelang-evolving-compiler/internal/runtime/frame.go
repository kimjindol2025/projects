package runtime

// CallFrame represents a function call frame in the call stack
type CallFrame struct {
	FnName string
	Locals map[string]Value
}

// NewFrame creates a new call frame for a function
// It binds the parameters to their corresponding arguments
func NewFrame(fnName string, params []string, args []Value) *CallFrame {
	frame := &CallFrame{
		FnName: fnName,
		Locals: make(map[string]Value),
	}

	// Bind parameters to arguments
	for i, paramName := range params {
		if i < len(args) {
			frame.Locals[paramName] = args[i]
		} else {
			frame.Locals[paramName] = NilVal()
		}
	}

	return frame
}

// Get retrieves a value from the frame's local variables
func (f *CallFrame) Get(name string) (Value, bool) {
	v, exists := f.Locals[name]
	return v, exists
}

// Set stores a value in the frame's local variables
func (f *CallFrame) Set(name string, v Value) {
	f.Locals[name] = v
}

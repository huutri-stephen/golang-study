package main

import (
	"fmt"
	"io"
	"strings"
)

func main() {
	fmt.Println("=== 1. Interface Basics ===")
	interfaceBasics()

	fmt.Println("\n=== 2. Nil Interface vs Typed Nil ===")
	nilInterfaceDemo()

	fmt.Println("\n=== 3. Method Set Rules ===")
	methodSetRules()

	fmt.Println("\n=== 4. Interface Composition ===")
	interfaceComposition()

	fmt.Println("\n=== 5. Type Assertion & Type Switch ===")
	typeAssertionDemo()

	fmt.Println("\n=== 6. Empty Interface ===")
	emptyInterfaceDemo()

	fmt.Println("\n=== 7. Interface Internals ===")
	interfaceInternals()
}

// --- Interfaces ---

type Animal interface {
	Speak() string
	Name() string
}

type Mover interface {
	Move() string
}

// Composition
type Pet interface {
	Animal
	Mover
	Owner() string
}

// --- Implementations ---

type Dog struct {
	name  string
	owner string
}

func (d *Dog) Speak() string { return "Woof!" }
func (d *Dog) Name() string  { return d.name }
func (d *Dog) Move() string  { return "runs" }
func (d *Dog) Owner() string { return d.owner }

type Cat struct {
	name string
}

func (c Cat) Speak() string { return "Meow!" } // value receiver
func (c Cat) Name() string  { return c.name }  // value receiver

func interfaceBasics() {
	// Pointer receiver → must use pointer
	var a Animal = &Dog{name: "Rex"}
	fmt.Printf("Dog: %s says %s\n", a.Name(), a.Speak())

	// Value receiver → can use both value and pointer
	var b Animal = Cat{name: "Whiskers"}
	fmt.Printf("Cat (value): %s says %s\n", b.Name(), b.Speak())

	var c Animal = &Cat{name: "Felix"}
	fmt.Printf("Cat (pointer): %s says %s\n", c.Name(), c.Speak())
}

func nilInterfaceDemo() {
	// Case 1: True nil interface
	var i interface{} = nil
	fmt.Printf("nil interface: value=%v, isNil=%v\n", i, i == nil) // true

	// Case 2: Typed nil — THE TRAP!
	var p *Dog = nil
	var a Animal = p                                           // interface has type (*Dog) but nil value
	fmt.Printf("typed nil: value=%v, isNil=%v\n", a, a == nil) // false!

	// This is a common bug in error handling:
	fmt.Println("\n--- Common Bug ---")
	err := getError(false)
	if err != nil {
		fmt.Printf("ERROR: err is not nil! type=%T, value=%v\n", err, err)
	}

	// Correct way:
	err2 := getErrorCorrect(false)
	if err2 != nil {
		fmt.Println("This won't print")
	} else {
		fmt.Println("CORRECT: err2 is nil")
	}
}

type MyError struct {
	msg string
}

func (e *MyError) Error() string { return e.msg }

// BAD: returns typed nil
func getError(fail bool) error {
	var err *MyError = nil
	if fail {
		err = &MyError{msg: "failed"}
	}
	return err // returns interface{type: *MyError, value: nil} — NOT nil!
}

// GOOD: returns untyped nil
func getErrorCorrect(fail bool) error {
	if fail {
		return &MyError{msg: "failed"}
	}
	return nil // returns interface{type: nil, value: nil} — IS nil
}

func methodSetRules() {
	fmt.Println(`
Method Set Rules:
┌─────────────────────────────────────────────────────┐
│ Type T  → method set includes: value receivers only │
│ Type *T → method set includes: value + pointer recv │
└─────────────────────────────────────────────────────┘
`)

	// Dog has pointer receivers → only *Dog implements Animal
	// var _ Animal = Dog{}   // COMPILE ERROR
	var _ Animal = &Dog{} // OK

	// Cat has value receivers → both Cat and *Cat implement Animal
	var _ Animal = Cat{}  // OK
	var _ Animal = &Cat{} // OK

	fmt.Println("Dog (pointer receiver): only &Dog{} implements Animal")
	fmt.Println("Cat (value receiver): both Cat{} and &Cat{} implement Animal")
}

func interfaceComposition() {
	dog := &Dog{name: "Buddy", owner: "Alice"}

	// Dog implements Pet (Animal + Mover + Owner())
	var pet Pet = dog
	fmt.Printf("Pet: %s, speaks: %s, moves: %s, owner: %s\n",
		pet.Name(), pet.Speak(), pet.Move(), pet.Owner())

	// Can assign to sub-interfaces
	var animal Animal = pet
	var mover Mover = pet
	fmt.Printf("As Animal: %s\n", animal.Speak())
	fmt.Printf("As Mover: %s\n", mover.Move())
}

func typeAssertionDemo() {
	var i interface{} = "hello"

	// Type assertion
	s, ok := i.(string)
	fmt.Printf("String assertion: %q, ok=%v\n", s, ok)

	n, ok := i.(int)
	fmt.Printf("Int assertion: %d, ok=%v\n", n, ok)

	// Type switch
	values := []interface{}{42, "hello", true, 3.14, nil}
	for _, v := range values {
		switch val := v.(type) {
		case int:
			fmt.Printf("  int: %d\n", val)
		case string:
			fmt.Printf("  string: %q\n", val)
		case bool:
			fmt.Printf("  bool: %v\n", val)
		case nil:
			fmt.Println("  nil")
		default:
			fmt.Printf("  unknown: %T\n", val)
		}
	}
}

func emptyInterfaceDemo() {
	// interface{} / any can hold anything
	var box interface{}

	box = 42
	fmt.Printf("int: %v (%T)\n", box, box)

	box = "hello"
	fmt.Printf("string: %v (%T)\n", box, box)

	box = struct{ X int }{X: 10}
	fmt.Printf("struct: %v (%T)\n", box, box)

	// Practical: io.Reader interface
	var r io.Reader = strings.NewReader("hello world")
	buf := make([]byte, 5)
	n, _ := r.Read(buf)
	fmt.Printf("Read %d bytes: %s\n", n, buf[:n])
}

func interfaceInternals() {
	fmt.Println(`
Interface Internal Representation:

1. Empty Interface (eface) — interface{} / any:
┌──────────────────────┐
│ _type  *_type        │ ← pointer to type descriptor
│ data   unsafe.Pointer│ ← pointer to actual value
└──────────────────────┘

2. Non-empty Interface (iface) — has methods:
┌──────────────────────┐
│ tab   *itab          │ ← interface table
│ data  unsafe.Pointer │ ← pointer to actual value
└──────────────────────┘

itab structure:
┌──────────────────────┐
│ inter  *interfacetype│ ← interface type info
│ _type  *_type        │ ← concrete type info  
│ hash   uint32        │ ← type hash (for fast type assertion)
│ fun    [1]uintptr    │ ← method table (variable length)
└──────────────────────┘

Key Insights:
• Interface value holds pointer to data (even for value types)
• Small values (≤ pointer size) may be stored directly
• itab is cached and reused for same interface+concrete type pair
• Method call: load itab → index into fun[] → indirect call
• Cost: ~2-3ns per interface method call vs direct call
• Interface prevents inlining of method calls
`)
}

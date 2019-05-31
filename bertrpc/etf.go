package bertrpc // import "gosrc.io/erlang/bertrpc"

// Supported ETF types
const (
	TagSmallInteger   = 97
	TagInteger        = 98
	TagDeprecatedAtom = 100
	TagSmallTuple     = 104
	TagLargeTuple     = 105
	TagNil            = 106
	TagString         = 107
	TagList           = 108
	TagBinary         = 109
	TagAtomUTF8       = 118
	TagSmallAtomUTF8  = 119
	TagETFVersion     = 131
)

// tagName convert a tag ID to its human readable tag name.
func tagName(tag int) string {
	switch tag {
	case TagSmallInteger:
		return "SmallInteger"
	case TagInteger:
		return "Integer"
	case TagDeprecatedAtom:
		return "DeprecatedAtom"
	case TagSmallTuple:
		return "SmallTuple"
	case TagLargeTuple:
		return "LargeTuple"
	case TagNil:
		return "Nil"
	case TagString:
		return "String"
	case TagList:
		return "List"
	case TagBinary:
		return "Binary"
	case TagAtomUTF8:
		return "AtomUTF8"
	case TagSmallAtomUTF8:
		return "SmallAtomUTF"
	case TagETFVersion:
		return "VersionTag"
	default:
		return string(tag)
	}
}

// ============================================================================
// String / Atom wrapper

type StringType int

const (
	StringTypeString = iota
	StringTypeAtom
)

// String is a wrapper structure to support Erlang atom or string data type.
// This type can be used when you want control / access to the underlying representation,
// for example to make a difference between atoms and binaries.
// If the difference does not matter for your code, you can simply use Go built-in string type.
type String struct {
	Value      string
	ErlangType StringType
}

func (str String) String() string {
	return str.Value
}

func (str String) IsAtom() bool {
	return str.ErlangType == StringTypeAtom
}

// ============================================================================
// List / Collection types

type Tuple struct {
	Elems []interface{}
}

type List []interface{}

// Charlist is a wrapper structure to support Erlang charlist in encoding.
// Charlist is only used in encoding. On decoding, charlists are always decoded
// as strings.
type CharList struct {
	Value string
}

// ============================================================================
// Helpers
// Short factory functions to help write short structure generation code.

// Atom
func A(atom string) String {
	return String{Value: atom, ErlangType: StringTypeAtom}
}

// String
func S(str string) String {
	return String{Value: str}
}

// Tuple
func T(el ...interface{}) Tuple {
	return Tuple{el}
}

// List
func L(el ...interface{}) []interface{} {
	return el
}

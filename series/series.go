package series

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"math"

	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"
)

// series is a data structure designed for operating on arrays of elements that
// should comply with a certain type structure. They are flexible enough that can
// be transformed to other series types and account for missing or non valid
// elements. Most of the power of series resides on the ability to compare and
// subset series of different types.
type series struct {
	name     string   // The name of the series
	elements Elements // The values of the elements
	t        Type     // The type of the series

	// deprecated: use Error() instead
	err error
}

// Elements is the interface that represents the array of elements contained on
// a Series.
type Elements interface {
	Elem(int) Element
	Len() int
	Slice(start, end int) Elements
	Get(indexs ...int) Elements
	Append(Elements) Elements
	AppendOne(Element) Elements
	Copy() Elements
}

// Element is the interface that defines the types of methods to be present for
// elements of a Series
type Element interface {
	// Setter method
	Set(interface{})
	SetElement(val Element)
	SetBool(val bool)
	SetFloat(val float64)
	SetInt(val int)
	SetString(val string)

	// Comparation methods
	Eq(Element) bool
	Neq(Element) bool
	Less(Element) bool
	LessEq(Element) bool
	Greater(Element) bool
	GreaterEq(Element) bool

	// Accessor/conversion methods
	Copy() Element     // FIXME: Returning interface is a recipe for pain
	Val() ElementValue // FIXME: Returning interface is a recipe for pain
	String() string
	Int() (int, error)
	Float() float64
	Bool() (bool, error)

	// Information methods
	IsNA() bool
	Type() Type
}

type Series interface {
	Rolling(window int, minPeriods int) RollingSeries
	// HasNaN checks whether the Series contain NaN elements.
	HasNaN() bool
	// IsNaN returns an array that identifies which of the elements are NaN.
	IsNaN() []bool
	// IsNotNaN returns an array that identifies which of the elements are not NaN.
	IsNotNaN() []bool
	// Compare compares the values of a Series with other elements. To do so, the
	// elements with are to be compared are first transformed to a Series of the same
	// type as the caller.
	Compare(comparator Comparator, comparando interface{}) Series
	// Float returns the elements of a Series as a []float64. If the elements can not
	// be converted to float64 or contains a NaN returns the float representation of
	// NaN.
	Float() []float64
	// Bool returns the elements of a Series as a []bool or an error if the
	// transformation is not possible.
	Bool() ([]bool, error)
	// Int returns the elements of a Series as a []int or an error if the
	// transformation is not possible.
	Int() ([]int, error)
	// Order returns the indexes for sorting a Series. NaN elements are pushed to the
	// end by order of appearance.
	Order(reverse bool) []int
	// StdDev calculates the standard deviation of a series
	StdDev() float64
	// Mean calculates the average value of a series
	Mean() float64
	// Median calculates the middle or median value, as opposed to
	// mean, and there is less susceptible to being affected by outliers.
	Median() float64
	// Max return the biggest element in the series
	Max() float64
	// MaxStr return the biggest element in a series of type String
	MaxStr() string
	// Min return the lowest element in the series
	Min() float64
	// MinStr return the lowest element in a series of type String
	MinStr() string
	// Quantile returns the sample of x such that x is greater than or
	// equal to the fraction p of samples.
	// Note: gonum/stat panics when called with strings
	Quantile(p float64) float64
	Quantiles(ps ...float64) []float64
	// DataQuantile returns the data quantile in the series
	DataQuantile(data float64) float64
	DataQuantiles(datas ...float64) []float64
	// Map applies a function matching MapFunction signature, which itself
	// allowing for a fairly flexible MAP implementation, intended for mapping
	// the function over each element in Series and returning a new Series object.
	// Function must be compatible with the underlying type of data in the Series.
	// In other words it is expected that when working with a Float Series, that
	// the function passed in via argument `f` will not expect another type, but
	// instead expects to handle Element(s) of type Float.
	Map(f MapFunction) Series
	//Shift series by desired number of periods and returning a new Series object.
	Shift(periods int) Series
	// CumProd finds the cumulative product of the first i elements in s and returning a new Series object.
	CumProd() Series
	// Prod returns the product of the elements of the Series. Returns 1 if len(s) = 0.
	Prod() float64
	// AddConst adds the scalar c to all of the values in Series and returning a new Series object.
	AddConst(c float64) Series
	// AddConst multiply the scalar c to all of the values in Series and returning a new Series object.
	MulConst(c float64) Series
	// DivConst Div the scalar c to all of the values in Series and returning a new Series object.
	DivConst(c float64) Series
	Add(c Series) Series
	Sub(c Series) Series
	Mul(c Series) Series
	Div(c Series) Series
	Abs() Series
	// Sum calculates the sum value of a series
	Sum() float64
	// Empty returns an empty Series of the same type

	Empty() Series
	// Returns Error or nil if no error occured
	Error() error
	// Subset returns a subset of the series based on the given Indexes.
	Subset(indexes Indexes) Series
	// Concat concatenates two series together. It will return a new Series with the
	// combined elements of both Series.
	Concat(x Series) Series
	// Copy will return a copy of the Series.
	Copy() Series
	// Records returns the elements of a Series as a []string
	Records() []string
	// Type returns the type of a given series
	Type() Type
	// Len returns the length of a given Series
	Len() int
	// String implements the Stringer interface for Series
	String() string
	// Str prints some extra information about a given series
	Str() string
	// Val returns the value of a series for the given index. Will panic if the index
	// is out of bounds.
	Val(i int) interface{}
	// Elem returns the element of a series for the given index. Will panic if the
	// index is out of bounds.
	// The index could be less than 0. When the index equals -1, Elem returns the last element of a series.
	Elem(i int) Element
	// Slice slices Series from start to end-1 index.
	Slice(start, end int) Series
	// FillNaN Fill NaN values using the specified value.
	FillNaN(value ElementValue)
	// FillNaNForward Fill NaN values using the last non-NaN value
	FillNaNForward()
	// FillNaNBackward fill NaN values using the next non-NaN value
	FillNaNBackward()
	// CacheAble returns a cacheable series and the returned series's calculation will be cached in case of repeate calculation.
	CacheAble() Series
	// Immutable returns an immutable series and the series can not be modified.
	Immutable() Series
	// Set sets the values on the indexes of a Series and returns the reference
	// for itself. The original Series is modified.
	Set(indexes Indexes, newvalues Series) Series
	// Append adds new elements to the end of the Series. When using Append, the
	// Series is modified in place.
	Append(values interface{})
	Name() string
	SetName(name string)
	SetErr(err error)
	//And logical operation
	And(in interface{}) Series
	//Or logical operation
	Or(in interface{}) Series
	//Not logical operation
	Not() Series

	//Wrap define special operations for multiple Series
	Wrap(ss ...Series) Wrapper
	//When define conditional computation
	When(whenF WhenFilterFunction) When

	//Filter Select the elements that match the FilterFunction
	Filter(ff FilterFunction) Series

	// All the operations on Self will influence the Series's content.
	Self() Self
}

// intElements is the concrete implementation of Elements for Int elements.
type intElements []intElement

func (e intElements) Len() int                      { return len(e) }
func (e intElements) Elem(i int) Element            { return &e[i] }
func (e intElements) Slice(start, end int) Elements { return e[start:end] }
func (e intElements) Get(indexs ...int) Elements {
	elements := make(intElements, len(indexs))
	for k, i := range indexs {
		elements[k] = e[i]
	}
	return elements
}
func (e intElements) Append(elements Elements) Elements {
	eles := elements.(intElements)
	ret := append(e, eles...)
	return ret
}
func (e intElements) AppendOne(element Element) Elements {
	ele := element.(*intElement)
	ret := append(e, *ele)
	return ret
}

func (e intElements) Copy() Elements {
	elements := make(intElements, len(e))
	copy(elements, e)
	return elements
}

// stringElements is the concrete implementation of Elements for String elements.
type stringElements []stringElement

func (e stringElements) Len() int                      { return len(e) }
func (e stringElements) Elem(i int) Element            { return &e[i] }
func (e stringElements) Slice(start, end int) Elements { return e[start:end] }
func (e stringElements) Get(indexs ...int) Elements {
	elements := make(stringElements, len(indexs))
	for k, i := range indexs {
		elements[k] = e[i]
	}
	return elements
}
func (e stringElements) Append(elements Elements) Elements {
	eles := elements.(stringElements)
	ret := append(e, eles...)
	return ret
}
func (e stringElements) AppendOne(element Element) Elements {
	ele := element.(*stringElement)
	ret := append(e, *ele)
	return ret
}
func (e stringElements) Copy() Elements {
	elements := make(stringElements, len(e))
	copy(elements, e)
	return elements
}

// floatElements is the concrete implementation of Elements for Float elements.
type floatElements []floatElement

func (e floatElements) Len() int                      { return len(e) }
func (e floatElements) Elem(i int) Element            { return &e[i] }
func (e floatElements) Slice(start, end int) Elements { return e[start:end] }
func (e floatElements) Get(indexs ...int) Elements {
	elements := make(floatElements, len(indexs))
	for k, i := range indexs {
		elements[k] = e[i]
	}
	return elements
}
func (e floatElements) Append(elements Elements) Elements {
	eles := elements.(floatElements)
	ret := append(e, eles...)
	return ret
}
func (e floatElements) AppendOne(element Element) Elements {
	ele := element.(*floatElement)
	ret := append(e, *ele)
	return ret
}
func (e floatElements) Copy() Elements {
	elements := make(floatElements, len(e))
	copy(elements, e)
	return elements
}

// boolElements is the concrete implementation of Elements for Bool elements.
type boolElements []boolElement

func (e boolElements) Len() int                      { return len(e) }
func (e boolElements) Elem(i int) Element            { return &e[i] }
func (e boolElements) Slice(start, end int) Elements { return e[start:end] }
func (e boolElements) Get(indexs ...int) Elements {
	elements := make(boolElements, len(indexs))
	for k, i := range indexs {
		elements[k] = e[i]
	}
	return elements
}
func (e boolElements) Append(elements Elements) Elements {
	eles := elements.(boolElements)
	ret := append(e, eles...)
	return ret
}
func (e boolElements) AppendOne(element Element) Elements {
	ele := element.(*boolElement)
	ret := append(e, *ele)
	return ret
}
func (e boolElements) Copy() Elements {
	elements := make(boolElements, len(e))
	copy(elements, e)
	return elements
}

// ElementValue represents the value that can be used for marshaling or
// unmarshaling Elements.
type ElementValue interface{}

type MapFunction func(ele Element, index int) Element

// Comparator is a convenience alias that can be used for a more type safe way of
// reason and use comparators.
type Comparator string

// Supported Comparators
const (
	Eq        Comparator = "=="   // Equal
	Neq       Comparator = "!="   // Non equal
	Greater   Comparator = ">"    // Greater than
	GreaterEq Comparator = ">="   // Greater or equal than
	Less      Comparator = "<"    // Lesser than
	LessEq    Comparator = "<="   // Lesser or equal than
	In        Comparator = "in"   // Inside
	CompFunc  Comparator = "func" // user-defined comparison function
)

// compFunc defines a user-defined comparator function. Used internally for type assertions
type compFunc = func(el Element) bool

// Type is a convenience alias that can be used for a more type safe way of
// reason and use Series types.
type Type string

// Supported Series Types
const (
	String Type = "string"
	Int    Type = "int"
	Float  Type = "float"
	Bool   Type = "bool"
)

func (t Type) emptyElements(n int) Elements {
	var elements Elements
	switch t {
	case String:
		elements = make(stringElements, n)
	case Int:
		elements = make(intElements, n)
	case Float:
		elements = make(floatElements, n)
	case Bool:
		elements = make(boolElements, n)
	default:
		panic(fmt.Sprintf("unknown type %v", t))
	}
	return elements
}

const NaN = "NaN"

// Indexes represent the elements that can be used for selecting a subset of
// elements within a Series. Currently supported are:
//
//     int            // Matches the given index number
//     []int          // Matches all given index numbers
//     []bool         // Matches all elements in a Series marked as true
//     Series [Int]   // Same as []int
//     Series [Bool]  // Same as []bool
type Indexes interface{}

var _ Series = (*series)(nil)

func Err(err error) Series {
	return &series{err: err}
}

// New is the generic Series constructor
func New(values interface{}, t Type, name string) Series {
	ret := newSeries(values, t, name)
	return &ret
}
func newSeries(values interface{}, t Type, name string) series {
	ret := series{
		name: name,
		t:    t,
	}

	// Pre-allocate elements
	preAlloc := func(n int) {
		ret.elements = t.emptyElements(n)
	}

	if values == nil {
		preAlloc(1)
		ret.elements.Elem(0).Set(nil)
		return ret
	}

	switch v := values.(type) {
	case []string:
		l := len(v)
		preAlloc(l)
		for i := 0; i < l; i++ {
			ret.elements.Elem(i).SetString(v[i])
		}
	case []float64:
		l := len(v)
		preAlloc(l)
		for i := 0; i < l; i++ {
			ret.elements.Elem(i).SetFloat(v[i])
		}
	case []int:
		l := len(v)
		preAlloc(l)
		for i := 0; i < l; i++ {
			ret.elements.Elem(i).SetInt(v[i])
		}
	case []bool:
		l := len(v)
		preAlloc(l)
		for i := 0; i < l; i++ {
			ret.elements.Elem(i).SetBool(v[i])
		}
	case []Element:
		l := len(v)
		preAlloc(l)
		for i := 0; i < l; i++ {
			ret.elements.Elem(i).SetElement(v[i])
		}
	case Series:
		l := v.Len()
		preAlloc(l)
		for i := 0; i < l; i++ {
			ret.elements.Elem(i).SetElement(v.Elem(i))
		}
	default:
		switch reflect.TypeOf(values).Kind() {
		case reflect.Slice:
			v := reflect.ValueOf(values)
			l := v.Len()
			preAlloc(v.Len())
			for i := 0; i < l; i++ {
				val := v.Index(i).Interface()
				ret.elements.Elem(i).Set(val)
			}
		default:
			preAlloc(1)
			v := reflect.ValueOf(values)
			val := v.Interface()
			ret.elements.Elem(0).Set(val)
		}
	}

	return ret
}

func NewDefault(defaultValue interface{}, t Type, name string, len int) Series {
	ret := &series{
		name: name,
		t:    t,
	}

	// Pre-allocate elements
	preAlloc := func(n int) {
		ret.elements = t.emptyElements(n)
	}

	if defaultValue == nil {
		preAlloc(1)
		ret.elements.Elem(0).Set(nil)
		return ret
	}
	preAlloc(len)

	switch v := defaultValue.(type) {
	case string:
		for i := 0; i < len; i++ {
			ret.elements.Elem(i).SetString(v)
		}
	case float64:
		for i := 0; i < len; i++ {
			ret.elements.Elem(i).SetFloat(v)
		}
	case int:
		for i := 0; i < len; i++ {
			ret.elements.Elem(i).SetInt(v)
		}
	case bool:
		for i := 0; i < len; i++ {
			ret.elements.Elem(i).SetBool(v)
		}
	case Element:
		for i := 0; i < len; i++ {
			ret.elements.Elem(i).SetElement(v)
		}
	default:
		for i := 0; i < len; i++ {
			ret.elements.Elem(i).Set(defaultValue)
		}
	}
	return ret
}

// Strings is a constructor for a String Series
func Strings(values interface{}) Series {
	return New(values, String, "")
}

// Ints is a constructor for an Int Series
func Ints(values interface{}) Series {
	return New(values, Int, "")
}

// Floats is a constructor for a Float Series
func Floats(values interface{}) Series {
	return New(values, Float, "")
}

// Bools is a constructor for a Bool Series
func Bools(values interface{}) Series {
	return New(values, Bool, "")
}

// Empty returns an empty Series of the same type
func (s series) Empty() Series {
	return New([]int{}, s.t, s.name)
}

// Returns Error or nil if no error occured
func (s *series) Error() error {
	return s.err
}

func (s *series) SetErr(err error) {
	s.err = err
}

// Append adds new elements to the end of the Series. When using Append, the
// Series is modified in place.
func (s *series) Append(values interface{}) {
	if err := s.err; err != nil {
		return
	}
	news := newSeries(values, s.t, s.name)
	s.elements = s.elements.Append(news.elements)
}

// Concat concatenates two series together. It will return a new Series with the
// combined elements of both Series.
func (s series) Concat(x Series) Series {
	if err := s.err; err != nil {
		return &s
	}
	if err := x.Error(); err != nil {
		s.err = fmt.Errorf("concat error: argument has errors: %v", err)
		return &s
	}
	y := s.Copy()
	y.Append(x)
	return y
}

// Subset returns a subset of the series based on the given Indexes.
func (s series) Subset(indexes Indexes) Series {
	if err := s.err; err != nil {
		return &s
	}
	idx, err := parseIndexes(s.Len(), indexes)
	if err != nil {
		s.err = err
		return &s
	}
	ret := &series{
		name:     s.name,
		t:        s.t,
		elements: s.elements.Get(idx...),
	}
	return ret
}

// Set sets the values on the indexes of a Series and returns the reference
// for itself. The original Series is modified.
func (s *series) Set(indexes Indexes, newvalues Series) Series {
	if err := s.err; err != nil {
		return s
	}
	if err := newvalues.Error(); err != nil {
		s.err = fmt.Errorf("set error: argument has errors: %v", err)
		return s
	}
	idx, err := parseIndexes(s.Len(), indexes)
	if err != nil {
		s.err = err
		return s
	}
	if len(idx) != newvalues.Len() {
		s.err = fmt.Errorf("set error: dimensions mismatch")
		return s
	}
	for k, i := range idx {
		if i < 0 || i >= s.Len() {
			s.err = fmt.Errorf("set error: index out of range")
			return s
		}
		s.elements.Elem(i).SetElement(newvalues.Elem(k))
	}
	return s
}

// HasNaN checks whether the Series contain NaN elements.
func (s series) HasNaN() bool {
	for i := 0; i < s.Len(); i++ {
		if s.elements.Elem(i).IsNA() {
			return true
		}
	}
	return false
}

// IsNaN returns an array that identifies which of the elements are NaN.
func (s series) IsNaN() []bool {
	ret := make([]bool, s.Len())
	for i := 0; i < s.Len(); i++ {
		ret[i] = s.elements.Elem(i).IsNA()
	}
	return ret
}

// IsNotNaN returns an array that identifies which of the elements are not NaN.
func (s series) IsNotNaN() []bool {
	ret := make([]bool, s.Len())
	for i := 0; i < s.Len(); i++ {
		ret[i] = !s.elements.Elem(i).IsNA()
	}
	return ret
}

// Compare compares the values of a Series with other elements. To do so, the
// elements with are to be compared are first transformed to a Series of the same
// type as the caller.
func (s series) Compare(comparator Comparator, comparando interface{}) Series {
	if err := s.err; err != nil {
		return &s
	}
	compareElements := func(a, b Element, c Comparator) (bool, error) {
		var ret bool
		switch c {
		case Eq:
			ret = a.Eq(b)
		case Neq:
			ret = a.Neq(b)
		case Greater:
			ret = a.Greater(b)
		case GreaterEq:
			ret = a.GreaterEq(b)
		case Less:
			ret = a.Less(b)
		case LessEq:
			ret = a.LessEq(b)
		default:
			return false, fmt.Errorf("unknown comparator: %v", c)
		}
		return ret, nil
	}

	bools := make([]bool, s.Len())

	// CompFunc comparator comparison
	if comparator == CompFunc {
		f, ok := comparando.(compFunc)
		if !ok {
			panic("comparando is not a comparison function of type func(el Element) bool")
		}

		for i := 0; i < s.Len(); i++ {
			e := s.elements.Elem(i)
			bools[i] = f(e)
		}

		return Bools(bools)
	}

	comp := newSeries(comparando, s.t, "")
	// In comparator comparison
	if comparator == In {
		for i := 0; i < s.Len(); i++ {
			e := s.elements.Elem(i)
			b := false
			for j := 0; j < comp.Len(); j++ {
				m := comp.elements.Elem(j)
				c, err := compareElements(e, m, Eq)
				if err != nil {
					s1 := s.Empty()
					s1.SetErr(err)
					return s1
				}
				if c {
					b = true
					break
				}
			}
			bools[i] = b
		}
		return Bools(bools)
	}

	// Single element comparison
	if comp.Len() == 1 {
		for i := 0; i < s.Len(); i++ {
			e := s.elements.Elem(i)
			c, err := compareElements(e, comp.elements.Elem(0), comparator)
			if err != nil {
				s1 := s.Empty()
				s1.SetErr(err)
				return s1
			}
			bools[i] = c
		}
		return Bools(bools)
	}

	// Multiple element comparison
	if s.Len() != comp.Len() {
		s1 := s.Empty()
		s1.SetErr(fmt.Errorf("can't compare: length mismatch"))
		return s1
	}
	for i := 0; i < s.Len(); i++ {
		e := s.elements.Elem(i)
		c, err := compareElements(e, comp.elements.Elem(i), comparator)
		if err != nil {
			s1 := s.Empty()
			s1.SetErr(err)
			return s1
		}
		bools[i] = c
	}
	return Bools(bools)
}

// Copy will return a copy of the Series.
func (s series) Copy() Series {
	ret := &series{
		name:     s.name,
		t:        s.t,
		elements: s.elements.Copy(),
		err:      s.err,
	}
	return ret
}

// Records returns the elements of a Series as a []string
func (s series) Records() []string {
	ret := make([]string, s.Len())
	for i := 0; i < s.Len(); i++ {
		e := s.elements.Elem(i)
		ret[i] = e.String()
	}
	return ret
}

// Float returns the elements of a Series as a []float64. If the elements can not
// be converted to float64 or contains a NaN returns the float representation of
// NaN.
func (s series) Float() []float64 {
	ret := make([]float64, s.Len())
	for i := 0; i < s.Len(); i++ {
		e := s.elements.Elem(i)
		ret[i] = e.Float()
	}
	return ret
}

// Int returns the elements of a Series as a []int or an error if the
// transformation is not possible.
func (s series) Int() ([]int, error) {
	ret := make([]int, s.Len())
	for i := 0; i < s.Len(); i++ {
		e := s.elements.Elem(i)
		val, err := e.Int()
		if err != nil {
			return nil, err
		}
		ret[i] = val
	}
	return ret, nil
}

// Bool returns the elements of a Series as a []bool or an error if the
// transformation is not possible.
func (s series) Bool() ([]bool, error) {
	ret := make([]bool, s.Len())
	for i := 0; i < s.Len(); i++ {
		e := s.elements.Elem(i)
		val, err := e.Bool()
		if err != nil {
			return nil, err
		}
		ret[i] = val
	}
	return ret, nil
}

// Type returns the type of a given series
func (s series) Type() Type {
	return s.t
}

// Len returns the length of a given Series
func (s series) Len() int {
	return s.elements.Len()
}

// String implements the Stringer interface for Series
func (s series) String() string {
	return fmt.Sprint(s.elements)
}

// Str prints some extra information about a given series
func (s series) Str() string {
	var ret []string
	// If name exists print name
	if s.name != "" {
		ret = append(ret, "Name: "+s.name)
	}
	ret = append(ret, "Type: "+fmt.Sprint(s.t))
	ret = append(ret, "Length: "+fmt.Sprint(s.Len()))
	if s.Len() != 0 {
		ret = append(ret, "Values: "+fmt.Sprint(s))
	}
	return strings.Join(ret, "\n")
}

// Val returns the value of a series for the given index. Will panic if the index
// is out of bounds.
func (s series) Val(i int) interface{} {
	return s.elements.Elem(i).Val()
}

// Elem returns the element of a series for the given index. Will panic if the
// index is out of bounds.
// The index could be less than 0. When the index equals -1, Elem returns the last element of a series.
func (s series) Elem(i int) Element {
	if i < 0 {
		return s.elements.Elem(s.Len() + i)
	}
	return s.elements.Elem(i)
}

// parseIndexes will parse the given indexes for a given series of length `l`. No
// out of bounds checks is performed.
func parseIndexes(l int, indexes Indexes) ([]int, error) {
	var idx []int
	switch idxs := indexes.(type) {
	case []int:
		idx = idxs
	case int:
		idx = []int{idxs}
	case []bool:
		bools := idxs
		if len(bools) != l {
			return nil, fmt.Errorf("indexing error: index dimensions mismatch")
		}
		for i, b := range bools {
			if b {
				idx = append(idx, i)
			}
		}
	case Series:
		s := idxs
		if err := s.Error(); err != nil {
			return nil, fmt.Errorf("indexing error: new values has errors: %v", err)
		}
		if s.HasNaN() {
			return nil, fmt.Errorf("indexing error: indexes contain NaN")
		}
		switch s.Type() {
		case Int:
			return s.Int()
		case Bool:
			bools, err := s.Bool()
			if err != nil {
				return nil, fmt.Errorf("indexing error: %v", err)
			}
			return parseIndexes(l, bools)
		default:
			return nil, fmt.Errorf("indexing error: unknown indexing mode")
		}
	default:
		return nil, fmt.Errorf("indexing error: unknown indexing mode")
	}
	return idx, nil
}

// Order returns the indexes for sorting a Series. NaN elements are pushed to the
// end by order of appearance.
func (s series) Order(reverse bool) []int {
	var ie indexedElements
	var nasIdx []int
	for i := 0; i < s.Len(); i++ {
		e := s.elements.Elem(i)
		if e.IsNA() {
			nasIdx = append(nasIdx, i)
		} else {
			ie = append(ie, indexedElement{i, e})
		}
	}
	var srt sort.Interface
	srt = ie
	if reverse {
		srt = sort.Reverse(srt)
	}
	sort.Stable(srt)
	var ret []int
	for _, e := range ie {
		ret = append(ret, e.index)
	}
	return append(ret, nasIdx...)
}

type indexedElement struct {
	index   int
	element Element
}

type indexedElements []indexedElement

func (e indexedElements) Len() int           { return len(e) }
func (e indexedElements) Less(i, j int) bool { return e[i].element.Less(e[j].element) }
func (e indexedElements) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }

// StdDev calculates the standard deviation of a series
func (s series) StdDev() float64 {
	stdDev := stat.StdDev(s.Float(), nil)
	return stdDev
}

// Mean calculates the average value of a series
func (s series) Mean() float64 {
	stdDev := stat.Mean(s.Float(), nil)
	return stdDev
}

// Median calculates the middle or median value, as opposed to
// mean, and there is less susceptible to being affected by outliers.
func (s series) Median() float64 {
	if s.elements.Len() == 0 ||
		s.Type() == String ||
		s.Type() == Bool {
		return math.NaN()
	}
	ix := s.Order(false)
	newElem := make([]Element, len(ix))

	for newpos, oldpos := range ix {
		newElem[newpos] = s.elements.Elem(oldpos)
	}

	// When length is odd, we just take length(list)/2
	// value as the median.
	if len(newElem)%2 != 0 {
		return newElem[len(newElem)/2].Float()
	}
	// When length is even, we take middle two elements of
	// list and the median is an average of the two of them.
	return (newElem[(len(newElem)/2)-1].Float() +
		newElem[len(newElem)/2].Float()) * 0.5
}

// Max return the biggest element in the series
func (s series) Max() float64 {
	if s.elements.Len() == 0 || s.Type() == String {
		return math.NaN()
	}

	max := s.elements.Elem(0)
	for i := 1; i < s.elements.Len(); i++ {
		elem := s.elements.Elem(i)
		if elem.Greater(max) {
			max = elem
		}
	}
	return max.Float()
}

// MaxStr return the biggest element in a series of type String
func (s series) MaxStr() string {
	if s.elements.Len() == 0 || s.Type() != String {
		return ""
	}

	max := s.elements.Elem(0)
	for i := 1; i < s.elements.Len(); i++ {
		elem := s.elements.Elem(i)
		if elem.Greater(max) {
			max = elem
		}
	}
	return max.String()
}

// Min return the lowest element in the series
func (s series) Min() float64 {
	if s.elements.Len() == 0 || s.Type() == String {
		return math.NaN()
	}

	min := s.elements.Elem(0)
	for i := 1; i < s.elements.Len(); i++ {
		elem := s.elements.Elem(i)
		if elem.Less(min) {
			min = elem
		}
	}
	return min.Float()
}

// MinStr return the lowest element in a series of type String
func (s series) MinStr() string {
	if s.elements.Len() == 0 || s.Type() != String {
		return ""
	}

	min := s.elements.Elem(0)
	for i := 1; i < s.elements.Len(); i++ {
		elem := s.elements.Elem(i)
		if elem.Less(min) {
			min = elem
		}
	}
	return min.String()
}

// Quantile returns the sample of x such that x is greater than or
// equal to the fraction p of samples.
// Note: gonum/stat panics when called with strings
func (s series) Quantile(p float64) float64 {
	if s.Type() == String || s.Len() == 0 {
		return math.NaN()
	}
	if p == 0 {
		return s.Min()
	}
	if p == 1 {
		return s.Max()
	}

	ordered := s.Subset(s.Order(false)).Float()

	return stat.Quantile(p, stat.Empirical, ordered, nil)
}

func (s series) Quantiles(ps ...float64) []float64 {
	if s.Type() == String || s.Len() == 0 {
		return nil
	}

	ret := make([]float64, len(ps))

	var ordered []float64
	for i := 0; i < len(ps); i++ {
		if ps[i] == 0 {
			ret[i] = s.Min()
			continue
		}
		if ps[i] == 1 {
			ret[i] = s.Max()
			continue
		}
		if ordered == nil {
			ordered = s.Subset(s.Order(false)).Float()
		}
		ret[i] = stat.Quantile(ps[i], stat.Empirical, ordered, nil)
	}

	return ret
}

// DataQuantile returns the data quantile in the series
func (s series) DataQuantile(data float64) float64 {
	if s.Type() == String || s.Len() == 0 {
		return math.NaN()
	}

	tmpS := s.Filter(func(ele Element, index int) bool {
		return !ele.IsNA()
	})
	if tmpS.Len() == 0 {
		return math.NaN()
	}

	ordered := tmpS.Subset(tmpS.Order(false)).Float()

	length := len(ordered)
	if length%2 == 1 {
		length = length + 1
	}

	ret := dataQuantile(data, ordered, length)
	return ret
}

func (s series) DataQuantiles(datas ...float64) []float64 {
	if s.Type() == String || s.Len() == 0 {
		return nil
	}

	tmpS := s.Filter(func(ele Element, index int) bool {
		return !ele.IsNA()
	})
	if tmpS.Len() == 0 {
		return nil
	}

	ordered := tmpS.Subset(tmpS.Order(false)).Float()

	length := len(ordered)
	if length%2 == 1 {
		length = length + 1
	}

	ret := make([]float64, len(datas))

	for j := 0; j < len(datas); j++ {
		ret[j] = dataQuantile(datas[j], ordered, length)
	}

	return ret
}

func dataQuantile(data float64, ordered []float64, length int) float64 {
	for i, d := range ordered {
		if data < d {
			return float64(i) / float64(length)
		}
	}
	return 1
}

// Map applies a function matching MapFunction signature, which itself
// allowing for a fairly flexible MAP implementation, intended for mapping
// the function over each element in Series and returning a new Series object.
// Function must be compatible with the underlying type of data in the Series.
// In other words it is expected that when working with a Float Series, that
// the function passed in via argument `f` will not expect another type, but
// instead expects to handle Element(s) of type Float.
func (s series) Map(f MapFunction) Series {
	eles := s.Type().emptyElements(s.Len())
	for i := 0; i < s.Len(); i++ {
		value := f(s.elements.Elem(i), i)
		eles.Elem(i).SetElement(value)
	}
	ret := &series{
		name:     s.name,
		elements: eles,
		t:        s.Type(),
		err:      nil,
	}
	return ret
}

//Shift series by desired number of periods and returning a new Series object.
func (s series) Shift(periods int) Series {
	if s.Len() == 0 {
		return s.Empty()
	}
	if periods == 0 {
		return s.Copy()
	}

	naLen := periods
	if naLen < 0 {
		naLen = -naLen
	}
	naEles := s.t.emptyElements(naLen)
	for i := 0; i < naLen; i++ {
		naEles.Elem(i).Set(NaN)
	}

	var shiftElements Elements
	if periods < 0 {
		//shift up
		shiftElements = s.elements.Slice(-periods, s.Len()).Copy().Append(naEles)
	} else if periods > 0 {
		//move down
		shiftElements = naEles.Append(s.elements.Slice(0, s.Len()-periods))
	}
	ret := &series{
		name:     fmt.Sprintf("%s_Shift(%d)", s.name, periods),
		elements: shiftElements,
		t:        s.t,
		err:      nil,
	}
	return ret
}

// CumProd finds the cumulative product of the first i elements in s and returning a new Series object.
func (s series) CumProd() Series {
	dst := make([]float64, s.Len())
	floats.CumProd(dst, s.Float())
	return New(dst, s.Type(), fmt.Sprintf("CumProd(%s)", s.name))
}

// Prod returns the product of the elements of the Series. Returns 1 if len(s) = 0.
func (s series) Prod() float64 {
	return floats.Prod(s.Float())
}

// AddConst adds the scalar c to all of the values in Series and returning a new Series object.
func (s series) AddConst(c float64) Series {
	dst := s.Float()
	floats.AddConst(c, dst)
	return New(dst, s.Type(), fmt.Sprintf("(%s + %v)", s.name, c))
}

// AddConst multiply the scalar c to all of the values in Series and returning a new Series object.
func (s series) MulConst(c float64) Series {
	sm := s.Map(func(e Element, index int) Element {
		result := e.Copy()
		f := result.Float()
		result.Set(f * c)
		return result
	})
	sm.SetName(fmt.Sprintf("(%s * %v)", s.name, c))
	return sm
}

// DivConst Div the scalar c to all of the values in Series and returning a new Series object.
func (s series) DivConst(c float64) Series {
	sm := s.Map(func(e Element, index int) Element {
		result := e.Copy()
		f := result.Float()
		result.Set(f / c)
		return result
	})
	sm.SetName(fmt.Sprintf("(%s / %v)", s.name, c))
	return sm
}

func (s series) Add(c Series) Series {
	sf := s.Float()
	cf := c.Float()
	dst := make([]float64, s.Len())
	floats.AddTo(dst, sf, cf)
	return New(dst, Float, fmt.Sprintf("(%s + %s)", s.name, c.Name()))
}

func (s series) Sub(c Series) Series {
	sf := s.Float()
	cf := c.Float()
	dst := make([]float64, s.Len())
	floats.SubTo(dst, sf, cf)
	return New(dst, Float, fmt.Sprintf("(%s - %s)", s.name, c.Name()))
}

func (s series) Mul(c Series) Series {
	sf := s.Float()
	cf := c.Float()
	dst := make([]float64, s.Len())
	floats.MulTo(dst, sf, cf)
	return New(dst, Float, fmt.Sprintf("(%s * %s)", s.name, c.Name()))
}

func (s series) Div(c Series) Series {
	sf := s.Float()
	cf := c.Float()
	dst := make([]float64, s.Len())
	floats.DivTo(dst, sf, cf)
	return New(dst, Float, fmt.Sprintf("(%s / %s)", s.name, c.Name()))
}

func (s series) Abs() Series {
	sm := s.Map(func(e Element, index int) Element {
		result := e.Copy()
		f := result.Float()
		result.Set(math.Abs(f))
		return result
	})
	sm.SetName(fmt.Sprintf("Abs(%s)", s.name))
	return sm
}

// FillNaN Fill NaN values using the specified value.
func (s series) FillNaN(value ElementValue) {
	for i := 0; i < s.Len(); i++ {
		ele := s.Elem(i)
		if ele.IsNA() {
			ele.Set(value)
		}
	}
}

// FillNaNForward Fill NaN values using the last non-NaN value
func (s series) FillNaNForward() {
	var lastNotNaNValue ElementValue = nil
	for i := 0; i < s.Len(); i++ {
		ele := s.Elem(i)
		if !ele.IsNA() {
			lastNotNaNValue = ele.Val()
		} else {
			if lastNotNaNValue != nil {
				ele.Set(lastNotNaNValue)
			}
		}
	}
}

// FillNaNBackward fill NaN values using the next non-NaN value
func (s series) FillNaNBackward() {
	var lastNotNaNValue ElementValue = nil
	for i := s.Len() - 1; i >= 0; i-- {
		ele := s.Elem(i)
		if !ele.IsNA() {
			lastNotNaNValue = ele.Val()
		} else {
			if lastNotNaNValue != nil {
				ele.Set(lastNotNaNValue)
			}
		}
	}
}

func (s series) Rolling(window int, minPeriods int) RollingSeries {
	return newRollingSeries(window, minPeriods, &s)
}

// CacheAble returns a cacheable series and the returned series's calculation will be cached in case of repeate calcution.
// You should make sure that the series will not be modified and has a unique name.
func (s series) CacheAble() Series {
	return newCacheAbleSeries(&s)
}
func (s series) Immutable() Series {
	return newImmutableSeries(&s)
}

//Operation for multiple series calculation
func Operation(operate func(index int, eles ...Element) interface{}, seriess ...Series) (Series, error) {
	if len(seriess) == 0 {
		return nil, errors.New("seriess num must > 0")
	}
	sl := seriess[0].Len()
	maxLen := sl
	for i := 1; i < len(seriess); i++ {
		slen := seriess[i].Len()
		if sl != slen && slen != 1 {
			return nil, errors.New("seriess length must be 1 or same")
		}
		if slen > maxLen {
			maxLen = slen
		}
	}

	t := seriess[0].Type()
	eles := t.emptyElements(maxLen)
	for i := 0; i < maxLen; i++ {
		operateParam := make([]Element, len(seriess))
		for j := 0; j < len(seriess); j++ {
			if seriess[j].Len() == 1 {
				operateParam[j] = seriess[j].Elem(0)
			} else {
				operateParam[j] = seriess[j].Elem(i)
			}
		}
		res := operate(i, operateParam...)
		eles.Elem(i).Set(res)
	}
	result := &series{
		name:     "",
		elements: eles,
		t:        t,
		err:      nil,
	}
	return result, nil
}

// Sum calculates the sum value of a series
func (s series) Sum() float64 {
	if s.elements.Len() == 0 || s.Type() == String {
		return math.NaN()
	}
	sFloat := s.Float()
	sum := sFloat[0]
	for i := 1; i < len(sFloat); i++ {
		sum += sFloat[i]
	}
	return sum
}

// Slice slices Series from start to end-1 index.
func (s series) Slice(start, end int) Series {
	if s.err != nil {
		return &s
	}

	if start > end || start < 0 || end > s.Len() {
		empty := s.Empty()
		empty.SetErr(fmt.Errorf("slice index out of bounds"))
		return empty
	}

	ret := &series{
		name: fmt.Sprintf("%s_Slice(%d,%d)", s.name, start, end),
		t:    s.t,
	}
	ret.elements = s.elements.Slice(start, end)
	return ret
}

func (s *series) SetName(name string) {
	s.name = name
}

func (s series) Name() string {
	return s.name
}

func (s *series) Wrap(ss ...Series) Wrapper {
	return newWrapper(s, ss)
}

func (s *series) When(whenF WhenFilterFunction) When {
	return newWhen(whenF, s)
}

//FilterFunction Select the elements that match the FilterFunction
type FilterFunction func(ele Element, index int) bool

func (s *series) Filter(ff FilterFunction) Series {
	eles := s.Type().emptyElements(0)
	for i := 0; i < s.Len(); i++ {
		ele := s.elements.Elem(i)
		if ff(ele, i) {
			eles = eles.AppendOne(ele)
		}
	}
	ret := &series{
		name:     s.name,
		elements: eles,
		t:        s.Type(),
		err:      nil,
	}
	return ret
}

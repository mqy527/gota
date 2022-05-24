package series

import (
	"fmt"
)

var _ Series = (*cacheAbleSeries)(nil)

type cacheAbleSeries struct {
	Series
	c Cache
}

func newCacheAbleSeries(s Series) Series {
	ret := &cacheAbleSeries{
		Series: s,
		c:      newSeriesCache(),
	}
	return ret
}

func (cs cacheAbleSeries) Rolling(window int, minPeriods int) RollingSeries {
	cr := cacheAbleRollingSeries{
		RollingSeries: NewRollingSeries(window, minPeriods, cs.Series),
		c:             newSeriesCache(),
	}
	return cr
}

func (cs cacheAbleSeries) HasNaN() bool {
	cacheKey := "HasNaN"
	ret, _ := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		ret := cs.Series.HasNaN()
		return ret, nil
	})
	return ret.(bool)
}

func (cs *cacheAbleSeries) cacheOrExecute(cacheKey string, f func() (interface{}, error)) (interface{}, error) {
	if ret, found := cs.c.Get(cacheKey); found {
		return ret, nil
	}
	ret, err := f()
	if err == nil {
		cs.c.Set(cacheKey, ret)
	}
	return ret, err
}

func (cs cacheAbleSeries) IsNaN() []bool {
	cacheKey := "IsNaN"
	ret, _ := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		ret := cs.Series.IsNaN()
		return ret, nil
	})
	return ret.([]bool)
}

func (cs cacheAbleSeries) IsNotNaN() []bool {
	cacheKey := "IsNotNaN"
	ret, _ := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		ret := cs.Series.IsNotNaN()
		return ret, nil
	})
	return ret.([]bool)
}

func (cs cacheAbleSeries) Float() []float64 {
	cacheKey := "Float"
	ret, _ := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		ret := cs.Series.Float()
		return ret, nil
	})
	return ret.([]float64)
}

func (cs cacheAbleSeries) Bool() ([]bool, error) {
	cacheKey := "Bool"
	ret, err := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		ret, err := cs.Series.Bool()
		return ret, err
	})
	return ret.([]bool), err
}

func (cs cacheAbleSeries) Int() ([]int, error) {
	cacheKey := "Int"
	ret, err := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		ret, err := cs.Series.Int()
		return ret, err
	})
	return ret.([]int), err
}

func (cs cacheAbleSeries) Records() []string {
	cacheKey := "Records"
	ret, _ := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		ret := cs.Series.Records()
		return ret, nil
	})
	return ret.([]string)
}

func (cs cacheAbleSeries) Order(reverse bool) []int {
	cacheKey := fmt.Sprintf("Order(%v)", reverse)
	ret, _ := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		ret := cs.Series.Order(reverse)
		return ret, nil
	})
	return ret.([]int)
}

func (cs cacheAbleSeries) StdDev() float64 {
	cacheKey := "StdDev"
	ret, _ := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		ret := cs.Series.StdDev()
		return ret, nil
	})
	return ret.(float64)
}

func (cs cacheAbleSeries) Mean() float64 {
	cacheKey := "Mean"
	ret, _ := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		ret := cs.Series.Mean()
		return ret, nil
	})
	return ret.(float64)
}

func (cs cacheAbleSeries) Median() float64 {
	cacheKey := "Median"
	ret, _ := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		ret := cs.Series.Median()
		return ret, nil
	})
	return ret.(float64)
}

func (cs cacheAbleSeries) Max() float64 {
	cacheKey := "Max"
	ret, _ := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		ret := cs.Series.Max()
		return ret, nil
	})
	return ret.(float64)
}

func (cs cacheAbleSeries) MaxStr() string {
	cacheKey := "MaxStr"
	ret, _ := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		ret := cs.Series.MaxStr()
		return ret, nil
	})
	return ret.(string)
}

func (cs cacheAbleSeries) Min() float64 {
	cacheKey := "Min"
	ret, _ := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		ret := cs.Series.Min()
		return ret, nil
	})
	return ret.(float64)
}

func (cs cacheAbleSeries) MinStr() string {
	cacheKey := "MinStr"
	ret, _ := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		ret := cs.Series.MinStr()
		return ret, nil
	})
	return ret.(string)
}

func (cs cacheAbleSeries) Quantile(p float64) float64 {
	cacheKey := fmt.Sprintf("Quantile(%f)", p)
	ret, _ := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		ret := cs.Series.Quantile(p)
		return ret, nil
	})
	return ret.(float64)
}

func (cs cacheAbleSeries) CumProd() Series {
	cacheKey := "CumProd"
	ret, _ := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		res := cs.Series.CumProd()
		ret := res.CacheAble()
		return ret, nil
	})
	return ret.(Series)
}

func (cs cacheAbleSeries) Prod() float64 {
	cacheKey := "Prod"
	ret, _ := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		ret := cs.Series.Prod()
		return ret, nil
	})
	return ret.(float64)
}

func (cs cacheAbleSeries) AddConst(c float64) Series {
	cacheKey := fmt.Sprintf("AddConst(%f)", c)
	ret, _ := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		res := cs.Series.AddConst(c)
		ret := res.CacheAble()
		return ret, nil
	})
	return ret.(Series)
}

func (cs cacheAbleSeries) MulConst(c float64) Series {
	cacheKey := fmt.Sprintf("MulConst(%f)", c)
	ret, _ := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		res := cs.Series.MulConst(c)
		ret := res.CacheAble()
		return ret, nil
	})
	return ret.(Series)
}

func (cs cacheAbleSeries) DivConst(c float64) Series {
	cacheKey := fmt.Sprintf("DivConst(%f)", c)
	ret, _ := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		res := cs.Series.DivConst(c)
		ret := res.CacheAble()
		return ret, nil
	})
	return ret.(Series)
}

func (cs cacheAbleSeries) Abs() Series {
	cacheKey := "Abs"
	ret, _ := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		res := cs.Series.Abs()
		ret := res.CacheAble()
		return ret, nil
	})
	return ret.(Series)
}

func (cs cacheAbleSeries) Not() Series {
	cacheKey := "Not"
	ret, _ := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		res := cs.Series.Not()
		ret := res.CacheAble()
		return ret, nil
	})
	return ret.(Series)
}

func (cs cacheAbleSeries) Sum() float64 {
	cacheKey := "Sum"
	ret, _ := cs.cacheOrExecute(cacheKey, func() (interface{}, error) {
		ret := cs.Series.Sum()
		return ret, nil
	})
	return ret.(float64)
}

func (cs cacheAbleSeries) Copy() Series {
	s := cs.Series.Copy()
	ret := &cacheAbleSeries{
		Series: s,
		c:      cs.c.Copy(),
	}
	return ret
}

func (cs *cacheAbleSeries) CacheAble() Series {
	return cs
}

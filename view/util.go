package view

import "fmt"

func Pluralize(count int, singular string, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

type Frac struct {
	num int
	den int
}

func NewFrac(num, den int) Frac {
	return Frac{num: num, den: den}
}

func (f Frac) String() string {
	return fmt.Sprintf("%d/%d", f.num, f.den)
}

func (f Frac) Percent() int {
	if f.den == 0 {
		return 0
	}
	return f.num * 100 / f.den
}

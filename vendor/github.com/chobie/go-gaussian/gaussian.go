package gaussian

// https://github.com/freethenation/gaussian
// free-gaussian
// MIT License

import (
	"math"
)

// prop
//mean: the mean (μ) of the distribution
//variance: the variance (σ^2) of the distribution
//standardDeviation: the standard deviation (σ) of the distribution

// combination
//mul(d): returns the product distribution of this and the given distribution. If a constant is passed in the distribution is scaled.
//div(d): returns the quotient distribution of this and the given distribution. If a constant is passed in the distribution is scaled by 1/d.
//add(d): returns the result of adding this and the given distribution
//sub(d): returns the result of subtracting this and the given distribution
//scale(c): returns the result of scaling this distribution by the given constant

type Gaussian struct {
	mean              float64
	variance          float64
	standardDeviation float64
}

func NewGaussian(mean, variance float64) *Gaussian {
	if variance <= 0.0 {
		panic("error")
	}

	return &Gaussian{
		mean:              mean,
		variance:          variance,
		standardDeviation: math.Sqrt(float64(variance)),
	}
}

// Complementary error function
// From Numerical Recipes in C 2e p221
func Erfc(x float64) float64 {
	z := math.Abs(x)
	t := 1 / (1 + z/2)
	r := t * math.Exp(-z*z-1.26551223+t*(1.00002368+
		t*(0.37409196+t*(0.09678418+t*(-0.18628806+
			t*(0.27886807+t*(-1.13520398+t*(1.48851587+
				t*(-0.82215223+t*0.17087277)))))))))
	if x >= 0 {
		return r
	} else {
		return 2 - r
	}
}

// Inverse complementary error function
// From Numerical Recipes 3e p265
func Ierfc(x float64) float64 {
	if x >= 2 {
		return -100
	}
	if x <= 0 {
		return 100
	}
	var xx float64
	if x < 1 {
		xx = x
	} else {
		xx = 2 - x
	}
	t := math.Sqrt(-2 * math.Log(xx/2))
	r := -0.70711 * ((2.30753+t*0.27061)/
		(1+t*(0.99229+t*0.04481)) - t)

	for j := 0; j < 2; j++ {
		e := Erfc(r) - xx
		r += e / (1.12837916709551257*math.Exp(-(r*r)) - r*e)
	}

	if x < 1 {
		return r
	} else {
		return -r
	}

}

// Construct a new distribution from the precision and precisionmean
func fromPrecisionMean(precision, precisionmean float64) *Gaussian {
	return NewGaussian(precisionmean/precision, 1/precision)
}

/// PROB

//pdf(x): the probability density function, which describes the probability
// of a random variable taking on the value x
func (self *Gaussian) Pdf(x float64) float64 {
	m := self.standardDeviation * math.Sqrt(2*math.Pi)
	e := math.Exp(-math.Pow(x-self.mean, 2) / (2 * self.variance))
	return e / m
}

//cdf(x): the cumulative distribution function,
// which describes the probability of a random
// variable falling in the interval (−∞, x]
func (self *Gaussian) Cdf(x float64) float64 {
	return 0.5 * Erfc(-(x-self.mean)/(self.standardDeviation*math.Sqrt(2)))
}

//ppf(x): the percent point function, the inverse of cdf
func (self *Gaussian) Ppf(x float64) float64 {
	return self.mean - self.standardDeviation*math.Sqrt(2)*Ierfc(2*x)
}

func (self *Gaussian) Add(d *Gaussian) *Gaussian {
	return NewGaussian(self.mean+d.mean, self.variance+d.variance)
}

func (self *Gaussian) Sub(d *Gaussian) *Gaussian {
	return NewGaussian(self.mean-d.mean, self.variance+d.variance)
}

func (self *Gaussian) Scale(c float64) *Gaussian {
	return NewGaussian(self.mean*c, self.variance*c*c)
}

func (self *Gaussian) Mul(d *Gaussian) *Gaussian {
	precision := 1 / self.variance
	dprecision := 1 / d.variance
	return fromPrecisionMean(precision+dprecision, precision*self.mean+dprecision*d.mean)
}

func (self *Gaussian) Div(d *Gaussian) *Gaussian {
	precision := 1 / self.variance
	dprecision := 1 / d.variance
	return fromPrecisionMean(precision-dprecision, precision*self.mean-dprecision*d.mean)
}

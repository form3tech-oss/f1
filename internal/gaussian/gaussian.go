package gaussian

import (
	"errors"
	"math"
)

const (
	Sqrt2Pi = 2.50662827463100050241576528481104525300698674060993831662992357 // https://oeis.org/A019727
)

// Distribution represents a Gaussian(Normal) distribution.
//
// The Gaussian distribution(also known as Normal distribution) is a bell-shaped curve that
// describes the probability of a random value occurring within a specific range.
//
//   - mean: the mean (μ) of the distribution (the centre of the bell curve)
//   - standardDeviation: the standard deviation (σ) of the distribution (the spread of the curve)
//   - variance: the variance (σ^2) of the distribution
type Distribution struct {
	mean              float64
	standardDeviation float64

	variance float64
}

// NewDistribution creates a new Gaussian instance.
//
// The Gaussian distribution(also known as Normal distribution) is a bell-shaped curve that
// describes the probability of a random value occurring within a specific range.
//
// It's characterized by two parameters:
//   - mean: the mean (μ) of the distribution (the centre of the bell curve)
//   - standardDeviation: the standard deviation (σ) of the distribution (the spread of the curve)
func NewDistribution(mean, standardDeviation float64) (*Distribution, error) {
	if standardDeviation <= 0.0 {
		return nil, errors.New("standard deviation must not be negative")
	}

	return &Distribution{
		mean:              mean,
		standardDeviation: standardDeviation,
		variance:          standardDeviation * standardDeviation,
	}, nil
}

// Exponent calculates the core Exponent term in the Gaussian distribution formula.
//
// This function is core to calculating the probability density and other related values.
//
// The core term of a Gaussian distribution(or normal distribution) is defined as:
//
//	e^(-(x - μ)^2 / (2σ^2))
//
// Where:
//   - x: The value at which to evaluate the Gaussian function.
//   - μ(mu): The mean of the distribution (the centre of the bell curve).
//   - σ(sigma): The standard deviation of the distribution (controls the spread of the curve).
//   - σ^2: The variance of the distribution
//   - e - Mathematical constant [math.E]
//
// This Exponent term determines the relative likelihood of a value `x` within the Gaussian
// distribution.
//
//   - The closer `x` to the mean `μ`, the closer the Exponent term gets to 1, indicating higher
//     probability density.
//   - As `x` moves away from the mean `μ`, the Exponent term decreases, indicating lower probability
//     density.
//   - The standard deviation `σ` controls the spread: Smaller standard deviation leads to sharper
//     curves with probability density concentrated around the mean.
func (d *Distribution) Exponent(x float64) float64 {
	pow2 := (x - d.mean) * (x - d.mean)
	return math.Exp(-pow2 / (2 * d.variance))
}

// PDF calculates the probability density function (PDF) at a given value.
//
// PDF of a Gaussian distribution is a bell-shaped curve, that describes the relative likelihood of
// a random variable taking on a specific value.
// Formula:
//
// f(x) = (1 / (σ * √(2π))) * e^(-(x - μ)^2 / (2σ^2))
//
// where:
//   - x: The value for which we want to calculate the probability density.
//   - μ(mu): The mean of the distribution (the centre of the bell curve).
//   - σ(sigma): The standard deviation of the distribution (controls the spread of the curve).
//   - σ^2: The variance of the distribution
//   - π - Mathematical constant [math.Pi]
//   - e - Mathematical constant [math.E]
func (d *Distribution) PDF(x float64) float64 {
	// Calculate the inverse of the normalisation constant for the normal distribution
	// The true normalisation constant is: (1 / (σ * √(2π)))
	m := d.standardDeviation * Sqrt2Pi

	return d.Exponent(x) / m
}

// CDF calculates the cumulative distribution function (CDF) of the Gaussian distribution.
//
// The CDF of a normal (Gaussian) distribution provides the probability of a random value (X) drawn
// from the distribution will be less than or equal to a specific value x.
//
// The CDF of a normal distribution with mean μ and standard deviation σ is typically denoted as:
// F(x) = P[X ≤ x]
//
// F(x) = 0.5 * (1 + erf((x - μ) / (σ * √2)))
// = 0.5 * erfc(-(x - μ) / (σ * √2))
//
// Where:
//   - x: The value at which to calculate the CDF
//   - σ(sigma): The standard deviation of the distribution.
//   - μ(mu): The mean of the distribution.
//   - erf(z): The error function
//   - erfc(z): The complementary error function([math.Erfc]) , erfc(z) = 1 - erf(z) ([math.Erf])
func (d *Distribution) CDF(x float64) float64 {
	// Transform normal distribution with mean μ and standard deviation σ
	// to a standard normal distribution with the following formula:
	//
	// z = (x - μ) / (σ * √2)
	//
	// Standard normal distribution is a special case of the normal distribution with a mean of 0
	// and a standard deviation of 1.
	z := (x - d.mean) / (d.standardDeviation * math.Sqrt2)

	// Calculate the CDF
	//
	// The complementary error function (erfc), denoted as erfc(z) is a special function defined as
	// "1 - erf(z)" , where erf(z) is the error function. The error function itself integrates the
	// probability density function(PDF) of the standard normal distribution. The erfc, on the other
	// hand, gives the probability that a standard normal random variable falls outside of the
	// interval [-z, z].
	//
	// By using the erfc function with the transformed input z, we can efficiently calculate the CDF
	// of the original normal distribution.
	//
	// Multiply by 0.5 because the Erfc function's output range is typically [0,1], while the CDF of
	// a normal distribution ranges from 0 (for negative infinity) to 1(for positive infinity)
	//
	// Use the Erfc(-z) because the Erfc function typically deals with non-negative arguments
	return 0.5 * math.Erfc(-z)
}

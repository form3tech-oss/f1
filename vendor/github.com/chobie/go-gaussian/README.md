# gaussian

A Golang model of the [Normal](http://en.wikipedia.org/wiki/Normal_distribution)
(or Gaussian) distribution. To install run `go get github.com/chobie/go-gaussian`

## API

### Creating a Distribution
```go
distribution := NewGausian(mean, variance)
```

### Properties
- `mean`: the mean (μ) of the distribution
- `variance`: the variance (σ^2) of the distribution
- `standardDeviation`: the standard deviation (σ) of the distribution

### Probability Functions
- `Pdf(x)`: the probability density function, which describes the probability
  of a random variable taking on the value _x_
- `Cdf(x)`: the cumulative distribution function, which describes the
  probability of a random variable falling in the interval (−∞, _x_]
- `Ppf(x)`: the percent point function, the inverse of _cdf_

### Combination Functions
- `Mul(d)`: returns the product distribution of this and the given distribution. If a constant is passed in the distribution is scaled.
- `Div(d)`: returns the quotient distribution of this and the given distribution. If a constant is passed in the distribution is scaled by 1/d.
- `Add(d)`: returns the result of adding this and the given distribution
- `Sub(d)`: returns the result of subtracting this and the given distribution
- `Scale(c)`: returns the result of scaling this distribution by the given constant

## History

This is a ported version of [freethenation's library](https://github.com/freethenation/gaussian)

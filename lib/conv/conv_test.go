package conv

import (
	"testing"

	"github.com/cockroachdb/apd/v2"
)

func Test_FromFloat(t *testing.T) {
	t.Run("should be able to convert a float into a decimal", func(t *testing.T) {
		res, err := FromFloat(123.321)
		if err != nil {
			t.Fatal(err)
		}
		if res.String() != "123.321" {
			t.Fatalf("expected 123.321, got %s", res.String())
		}
	})
}

func Test_FromString(t *testing.T) {
	t.Run("should be able to convert a string into a decimal", func(t *testing.T) {
		res, err := FromString("123.321")
		if err != nil {
			t.Fatal(err)
		}
		if res.String() != "123.321" {
			t.Fatalf("expected 123.321, got %s", res.String())
		}
	})
}

func TestMultplyDecimals(t *testing.T) {
	t.Run("should be able to multply two decimal values together", func(t *testing.T) {
		apdContext := apd.Context{
			Precision:   65,
			MaxExponent: 65,
			MinExponent: -18,
		}
		// from priceHr
		fromStringDecimal, err := FromString("10.246")
		if err != nil {
			t.Fatal(err)
		}
		// from float64 hours
		fromFloatDecimal, err := FromFloat(1.0123)
		if err != nil {
			t.Fatal(err)
		}
		// result decimal
		res := apd.New(65, -18)
		cond, err := apdContext.Mul(res, &fromStringDecimal, &fromFloatDecimal)
		if err != nil {
			t.Fatal(err)
		}
		if cond.Any() {
			t.Fatal("error")
		}
		if res.String() != "10.3720258" {
			t.Fatalf("expected 10.3720258, got %s", res.String())
		}
	})
}

func TestAddMultipleDecimals(t *testing.T) {
	decimalSlice := make([]apd.Decimal, 3)
	// decimal 1
	apdContext := apd.Context{
		Precision:   65,
		MaxExponent: 65,
		MinExponent: -18,
	}
	// from priceHr
	fromStringDecimal1, err := FromString("10.246")
	if err != nil {
		t.Fatal(err)
	}
	// from float64 hours
	fromFloatDecimal1, err := FromFloat(1.0123)
	if err != nil {
		t.Fatal(err)
	}
	// result decimal
	res1 := apd.New(65, -18)
	cond, err := apdContext.Mul(res1, &fromStringDecimal1, &fromFloatDecimal1)
	if err != nil {
		t.Fatal(err)
	}
	if cond.Any() {
		t.Fatal("error")
	}
	if res1.String() != "10.3720258" {
		t.Fatalf("expected 10.3720258, got %s", res1.String())
	}
	decimalSlice = append(decimalSlice, *res1)

	// decimal 2
	// from priceHr
	fromStringDecimal2, err := FromString("10.246")
	if err != nil {
		t.Fatal(err)
	}
	// from float64 hours
	fromFloatDecimal2, err := FromFloat(1.0123)
	if err != nil {
		t.Fatal(err)
	}
	// result decimal
	res2 := apd.New(65, -18)
	cond, err = apdContext.Mul(res2, &fromStringDecimal2, &fromFloatDecimal2)
	if err != nil {
		t.Fatal(err)
	}
	if cond.Any() {
		t.Fatal("error")
	}
	if res2.String() != "10.3720258" {
		t.Fatalf("expected 10.3720258, got %s", res2.String())
	}
	decimalSlice = append(decimalSlice, *res2)

	// decimal 3
	fromStringDecimal3, err := FromString("10.246")
	if err != nil {
		t.Fatal(err)
	}
	// from float64 hours
	fromFloatDecimal3, err := FromFloat(1.0123)
	if err != nil {
		t.Fatal(err)
	}
	// result decimal
	res3 := apd.New(65, -18)
	cond, err = apdContext.Mul(res3, &fromStringDecimal3, &fromFloatDecimal3)
	if err != nil {
		t.Fatal(err)
	}
	if cond.Any() {
		t.Fatal("error")
	}
	if res3.String() != "10.3720258" {
		t.Fatalf("expected 10.3720258, got %s", res1.String())
	}
	decimalSlice = append(decimalSlice, *res3)

	// calculate total of all three
	res := apd.New(65, -18)
	for _, d := range decimalSlice {
		cond, err = apdContext.Add(res, res, &d)
		if err != nil {
			t.Fatal(err)
		}
		if cond.Any() {
			t.Fatal("error")
		}
	}
	if res.String() != "31.116077400000000065" {
		t.Fatalf("expected 31.116077400000000065, got %s", res.String())
	}
}

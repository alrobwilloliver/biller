package conv

import (
	"fmt"

	"github.com/cockroachdb/apd/v2"
)

func FromFloat(val float64) (apd.Decimal, error) {
	hoursString := fmt.Sprintf("%v", val)
	res, err := FromString(hoursString)
	if err != nil {
		return apd.Decimal{}, fmt.Errorf("failure converting value: %w", err)
	}
	return res, nil
}

func FromString(val string) (apd.Decimal, error) {
	value, cond, err := apd.NewFromString(val)
	if err != nil {
		return apd.Decimal{}, fmt.Errorf("could not marshal value into decimal: %w", err)
	}
	if !cond.Any() {
		return *value, nil
	} else {
		_, err = cond.GoError(apd.DefaultTraps)
		return apd.Decimal{}, fmt.Errorf("could not marshal value into decimal: %w", err)
	}
}

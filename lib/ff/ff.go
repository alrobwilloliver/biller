package ff

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var envVarReplacer = strings.NewReplacer(
	"-", "_",
	".", "_",
	"/", "_",
)

// Parse the flags into the flag set from commandline args and environment variables
// in that priority order (command line args > environment variables).
func Fill(fs *flag.FlagSet, args []string) error {
	return FillUsingPrefix(fs, args, "")
}

func FillUsingPrefix(fs *flag.FlagSet, args []string, envVarPrefix string) error {
	err := fs.Parse(args)
	if err != nil {
		return err
	}

	provided := map[string]bool{}

	// visit all the flags that have been set
	fs.Visit(func(f *flag.Flag) {
		provided[f.Name] = true
	})

	// visit all the flags regardless if they have been set
	var visitErr error
	fs.VisitAll(func(f *flag.Flag) {
		if visitErr != nil {
			return
		}

		// if it has already been provided, then don't override it
		if provided[f.Name] {
			return
		}

		key := strings.ToUpper(f.Name)
		key = envVarReplacer.Replace(key)
		if envVarPrefix != "" {
			key = strings.ToUpper(envVarPrefix) + "_" + key
		}
		value := os.Getenv(key)
		if value == "" {
			return
		}

		if err := fs.Set(f.Name, value); err != nil {
			visitErr = fmt.Errorf("error setting flag %q from env var %q: %w", f.Name, key, err)
			return
		}
	})

	if visitErr != nil {
		return fmt.Errorf("error parsing env vars: %w", visitErr)
	}
	return nil
}

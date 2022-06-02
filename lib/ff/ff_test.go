package ff

import (
	"flag"
	"os"
	"testing"
)

func Test_FillEmpty(t *testing.T) {
	fs := flag.NewFlagSet("flaggy mcflagface", flag.ContinueOnError)
	args := []string{"arg1", "arg2"}

	if err := Fill(fs, args); err != nil {
		t.Errorf("TestFillEmpty error: %v", err)
	}
}

func Test_FillSingleCommandLineFlag(t *testing.T) {
	fs := flag.NewFlagSet("flaggy mcflagface", flag.ContinueOnError)
	args := []string{"--test-flag", "123"}
	var testFlag string
	fs.StringVar(&testFlag, "test-flag", "empty", "")

	err := Fill(fs, args)
	if err != nil {
		t.Errorf("TestFillSingleCommandLineFlag error: %v", err)
	}
	if testFlag != "123" {
		t.Errorf("expected: %v, actual: %v", "123", testFlag)
	}
}

func Test_FillWithPrecedence(t *testing.T) {
	os.Setenv("TEST_FLAG", "env-flag")
	fs := flag.NewFlagSet("flaggy mcflagface", flag.ContinueOnError)
	args := []string{"--test-flag", "cli-flag"}
	var testFlag string
	fs.StringVar(&testFlag, "test-flag", "empty", "")

	err := Fill(fs, args)
	if err != nil {
		t.Errorf("TestFillSingleCommandLineFlag error: %v", err)
	}
	if testFlag != "cli-flag" {
		t.Errorf("expected: cli-flag, actual: %v", testFlag)
	}
	os.Unsetenv("TEST_FLAG")
}

func Test_FillWithEnv(t *testing.T) {
	os.Setenv("TEST_FLAG", "env-flag")
	fs := flag.NewFlagSet("flaggy mcflagface", flag.ContinueOnError)
	args := []string{""}
	var testFlag string
	fs.StringVar(&testFlag, "test-flag", "empty", "")

	err := Fill(fs, args)
	if err != nil {
		t.Errorf("TestFillSingleCommandLineFlag error: %v", err)
	}
	if testFlag != "env-flag" {
		t.Errorf("expected: env-flag, actual: %v", testFlag)
	}
	os.Unsetenv("TEST_FLAG")
}

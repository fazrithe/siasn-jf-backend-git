package config_test

import (
	"fmt"
	"github.com/if-itb/siasn-libs-backend/config"
	"log"
	"os"
	"reflect"
	"testing"
)

func TestServiceConfig_ParseTo(t *testing.T) {
	type TestConfig struct {
		TestInt         int      `config:"TEST_INT"`
		TestString      string   `config:"TEST_STRING"`
		TestBool        bool     `config:"TEST_BOOL"`
		TestFloat32     float32  `config:"TEST_FLOAT32"`
		TestStringArray []string `config:"TEST_STRING_ARRAY"`
	}

	expect := &TestConfig{
		TestInt:         1,
		TestString:      "test",
		TestBool:        true,
		TestFloat32:     1.344,
		TestStringArray: []string{"abc", "cde"},
	}

	sc := config.ServiceConfig{
		Prefix:         "ABC",
		ArraySeparator: " ",
	}

	err := os.Setenv("ABC_TEST_INT", fmt.Sprint(expect.TestInt))
	if err != nil {
		t.Fatal(err)
	}

	err = os.Setenv("ABC_TEST_STRING", expect.TestString)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Setenv("ABC_TEST_BOOL", fmt.Sprint(expect.TestBool))
	if err != nil {
		t.Fatal(err)
	}

	err = os.Setenv("ABC_TEST_FLOAT32", fmt.Sprintf("%.3f", expect.TestFloat32))
	if err != nil {
		t.Fatal(err)
	}

	err = os.Setenv("ABC_TEST_STRING_ARRAY", fmt.Sprintf("%s %s", expect.TestStringArray[0], expect.TestStringArray[1]))
	if err != nil {
		t.Fatal(err)
	}

	n := &TestConfig{}
	err = sc.ParseTo(n)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expect, n) {
		t.Fatalf("decoded config is not the same with expectation, received: %v, expected: %v", n, expect)
	}
}

func ExampleServiceConfig_ParseTo() {
	type MyConfig struct {
		Port         int      `config:"PORT"`
		Host         string   `config:"HOST"`
		OtherConfigs []string `config:"OTHER_CONFIGS"`
		OtherValue   float32  `config:"OTHER_VALUE"`
	}

	sc := config.ServiceConfig{
		Prefix:         "MYSERVICE",
		ArraySeparator: " ",
	}

	_ = os.Setenv("MYSERVICE_PORT", "80")
	_ = os.Setenv("MYSERVICE_HOST", "192.168.1.1")
	_ = os.Setenv("MYSERVICE_OTHER_CONFIGS", "test test1")
	_ = os.Setenv("MYSERVICE_OTHER_VALUE", "1.234")

	myConfig := &MyConfig{}
	err := sc.ParseTo(myConfig)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(myConfig)
	// Output: &{80 192.168.1.1 [test test1] 1.234}
}

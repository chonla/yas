package testsuite

import (
	"fmt"
	"strings"

	"github.com/chonla/yas/assertable"
	"github.com/chonla/yas/request"
	"github.com/chonla/yas/response"
	"github.com/fatih/color"
)

// TestCase holds a test case
type TestCase struct {
	Name         string
	Method       string
	BaseURL      string
	Path         string
	ContentType  string
	RequestBody  string
	Headers      map[string]string
	Expectations map[string]string
	Captures     map[string]string
	Setups       []*Task
	Teardowns    []*Task
	Variables    map[string]string
}

// NewTestCase creates a new testcase
func NewTestCase(name string) *TestCase {
	return &TestCase{
		Name:         name,
		Headers:      map[string]string{},
		Expectations: map[string]string{},
		Captures:     map[string]string{},
		Setups:       []*Task{},
		Teardowns:    []*Task{},
		Variables:    map[string]string{},
	}
}

// SetContentType set a corresponding content type
func (tc *TestCase) SetContentType(ct string) {
	switch strings.ToLower(ct) {
	case "json":
		ct = "application/json"
	default:
		ct = "application/json"
	}
	tc.ContentType = ct
}

// Run executes test case
func (tc *TestCase) Run() error {
	// Skip if no assertion
	if len(tc.Expectations) == 0 {
		return nil
	}

	white := color.New(color.FgHiWhite, color.Bold).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	url := fmt.Sprintf("%s%s", tc.BaseURL, tc.Path)

	fmt.Printf("%s\n", white("================================================================================"))
	fmt.Printf("Testcase: %s\n", white(tc.Name))
	fmt.Printf("%s\n", white("================================================================================"))

	if len(tc.Setups) > 0 {
		for _, s := range tc.Setups {
			s.BaseURL = tc.BaseURL
			s.MergeVariables(tc.Variables)
			e := s.Run()
			if e != nil {
				fmt.Printf("%s: %s\n", red("Error"), e)
				return e
			}

			for k, v := range s.Captured {
				tc.Variables[k] = v
			}
		}
	}

	req, e := request.NewRequester(tc.Method)
	if e != nil {
		return e
	}

	req.SetHeaders(tc.applyVarsToMap(tc.Headers))
	resp, e := req.Request(tc.applyVars(url), tc.applyVars(tc.RequestBody))
	if e != nil {
		fmt.Printf("%s: %s\n", red("Error"), e)
		return e
	}

	as, e := assertable.NewAssertable(response.NewResponse(resp))
	if e != nil {
		return e
	}

	e = as.Assert(tc.Expectations)
	if e != nil {
		return e
	}

	if len(tc.Teardowns) > 0 {
		for _, s := range tc.Teardowns {
			s.BaseURL = tc.BaseURL
			s.MergeVariables(tc.Variables)
			e := s.Run()
			if e != nil {
				fmt.Printf("%s: %s\n", red("Error"), e)
				return e
			}

			for k, v := range s.Captured {
				tc.Variables[k] = v
			}
		}
	}

	return e
}

func (tc *TestCase) applyVarsToMap(data map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range data {
		out[k] = tc.applyVars(v)
	}
	return out
}

func (tc *TestCase) applyVars(data string) string {
	for k, v := range tc.Variables {
		data = strings.Replace(data, fmt.Sprintf("{%s}", k), v, -1)
	}
	return data
}

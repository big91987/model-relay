package relay

import "testing"

func TestCompletionAssistantRoleContract(test *testing.T) {
	for _, scenario := range []struct {
		name, body string
		valid      bool
		known      bool
	}{
		{"missing_role_known", `{"choices":[{"message":{"content":"controlled"}}],"usage":{"prompt_tokens":3,"completion_tokens":4,"total_tokens":7}}`, false, true},
		{"missing_role_unknown", `{"choices":[{"message":{"content":"controlled"}}]}`, false, false},
		{"wrong_role", `{"choices":[{"message":{"role":"user","content":"controlled"}}]}`, false, false},
		{"assistant_known", `{"choices":[{"message":{"role":"assistant","content":"controlled"}}],"usage":{"prompt_tokens":3,"completion_tokens":4,"total_tokens":7}}`, true, true},
		{"assistant_unknown", `{"choices":[{"message":{"role":"assistant","content":"controlled"}}]}`, true, false},
	} {
		test.Run(scenario.name, func(test *testing.T) {
			value, valid := validResponse([]byte(scenario.body), false)
			if valid != scenario.valid {
				test.Fatal("completion assistant role contract mismatch")
			}
			if valid {
				usage := usageFrom(value)
				if (usage.Prompt != nil && usage.Completion != nil) != scenario.known {
					test.Fatal("unknown usage must not become known")
				}
			}
		})
	}
}

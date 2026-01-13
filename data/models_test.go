package data_test

import (
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zidane0000/ai-interview-platform/data"
)

// Test language validation functions
func TestValidateLanguage(t *testing.T) {
	tests := []struct {
		name     string
		language string
		expected bool
	}{
		{"valid english", data.LanguageEnglish, true},
		{"valid traditional chinese", data.LanguageTraditionalChinese, true},
		{"invalid language", "fr", false},
		{"empty string", "", false},
		{"uppercase", "EN", false},
		{"mixed case", "En", false},
		{"invalid code", "zh-CN", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := data.ValidateLanguage(tt.language)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDefaultLanguage(t *testing.T) {
	result := data.GetDefaultLanguage()
	assert.Equal(t, data.LanguageEnglish, result)
}

func TestGetValidatedLanguage(t *testing.T) {
	tests := []struct {
		name     string
		language string
		expected string
	}{
		{"valid english", data.LanguageEnglish, data.LanguageEnglish},
		{"valid traditional chinese", data.LanguageTraditionalChinese, data.LanguageTraditionalChinese},
		{"invalid language defaults to english", "fr", data.LanguageEnglish},
		{"empty string defaults to english", "", data.LanguageEnglish},
		{"uppercase defaults to english", "EN", data.LanguageEnglish},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := data.GetValidatedLanguage(tt.language)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test interview type validation functions
func TestValidateInterviewType(t *testing.T) {
	tests := []struct {
		name          string
		interviewType string
		expected      bool
	}{
		{"valid general", data.InterviewTypeGeneral, true},
		{"valid technical", data.InterviewTypeTechnical, true},
		{"valid behavioral", data.InterviewTypeBehavioral, true},
		{"invalid type", "invalid", false},
		{"empty string", "", false},
		{"uppercase", "GENERAL", false},
		{"mixed case", "General", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := data.ValidateInterviewType(tt.interviewType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDefaultInterviewType(t *testing.T) {
	result := data.GetDefaultInterviewType()
	assert.Equal(t, data.InterviewTypeGeneral, result)
}

func TestGetValidatedInterviewType(t *testing.T) {
	tests := []struct {
		name          string
		interviewType string
		expected      string
	}{
		{"valid general", data.InterviewTypeGeneral, data.InterviewTypeGeneral},
		{"valid technical", data.InterviewTypeTechnical, data.InterviewTypeTechnical},
		{"valid behavioral", data.InterviewTypeBehavioral, data.InterviewTypeBehavioral},
		{"invalid type defaults to general", "invalid", data.InterviewTypeGeneral},
		{"empty string defaults to general", "", data.InterviewTypeGeneral},
		{"uppercase defaults to general", "GENERAL", data.InterviewTypeGeneral},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := data.GetValidatedInterviewType(tt.interviewType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test StringArray custom type
func TestStringArray_Scan(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expected    data.StringArray
		expectError bool
	}{
		{
			name:        "nil value",
			input:       nil,
			expected:    nil,
			expectError: false,
		},
		{
			name:        "valid json bytes",
			input:       []byte(`["item1", "item2", "item3"]`),
			expected:    data.StringArray{"item1", "item2", "item3"},
			expectError: false,
		},
		{
			name:        "valid json string",
			input:       `["item1", "item2"]`,
			expected:    data.StringArray{"item1", "item2"},
			expectError: false,
		},
		{
			name:        "empty array bytes",
			input:       []byte(`[]`),
			expected:    data.StringArray{},
			expectError: false,
		},
		{
			name:        "empty array string",
			input:       `[]`,
			expected:    data.StringArray{},
			expectError: false,
		},
		{
			name:        "invalid json bytes",
			input:       []byte(`["unclosed array"`),
			expected:    nil,
			expectError: true,
		},
		{
			name:        "invalid json string",
			input:       `["unclosed array"`,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "unsupported type",
			input:       123,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var arr data.StringArray
			err := arr.Scan(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, arr)
			}
		})
	}
}

func TestStringArray_Value(t *testing.T) {
	tests := []struct {
		name        string
		input       data.StringArray
		expected    driver.Value
		expectError bool
	}{
		{
			name:        "nil array",
			input:       nil,
			expected:    nil,
			expectError: false,
		},
		{
			name:        "empty array",
			input:       data.StringArray{},
			expected:    `[]`,
			expectError: false,
		},
		{
			name:        "single item array",
			input:       data.StringArray{"item1"},
			expected:    `["item1"]`,
			expectError: false,
		},
		{
			name:        "multiple items array",
			input:       data.StringArray{"item1", "item2", "item3"},
			expected:    `["item1","item2","item3"]`,
			expectError: false,
		},
		{
			name:        "array with special characters",
			input:       data.StringArray{"item with spaces", "item\"with\"quotes", "item\nwith\nnewlines"},
			expected:    `["item with spaces","item\"with\"quotes","item\nwith\nnewlines"]`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.input.Value()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expected == nil {
					assert.Nil(t, value)
				} else {
					// Convert []byte result to string for comparison
					if bytesValue, ok := value.([]byte); ok {
						assert.Equal(t, tt.expected, string(bytesValue))
					} else {
						assert.Equal(t, tt.expected, value)
					}
				}
			}
		})
	}
}

// Test StringMap custom type
func TestStringMap_Scan(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expected    data.StringMap
		expectError bool
	}{
		{
			name:        "nil value",
			input:       nil,
			expected:    nil,
			expectError: false,
		},
		{
			name:        "valid json bytes",
			input:       []byte(`{"key1": "value1", "key2": "value2"}`),
			expected:    data.StringMap{"key1": "value1", "key2": "value2"},
			expectError: false,
		},
		{
			name:        "valid json string",
			input:       `{"key1": "value1"}`,
			expected:    data.StringMap{"key1": "value1"},
			expectError: false,
		},
		{
			name:        "empty map bytes",
			input:       []byte(`{}`),
			expected:    data.StringMap{},
			expectError: false,
		},
		{
			name:        "empty map string",
			input:       `{}`,
			expected:    data.StringMap{},
			expectError: false,
		},
		{
			name:        "invalid json bytes",
			input:       []byte(`{"unclosed": "object"`),
			expected:    nil,
			expectError: true,
		},
		{
			name:        "invalid json string",
			input:       `{"unclosed": "object"`,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "unsupported type",
			input:       123,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m data.StringMap
			err := m.Scan(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, m)
			}
		})
	}
}

func TestStringMap_Value(t *testing.T) {
	tests := []struct {
		name        string
		input       data.StringMap
		expectError bool
	}{
		{
			name:        "nil map",
			input:       nil,
			expectError: false,
		},
		{
			name:        "empty map",
			input:       data.StringMap{},
			expectError: false,
		},
		{
			name:        "single key map",
			input:       data.StringMap{"key1": "value1"},
			expectError: false,
		},
		{
			name:        "multiple keys map",
			input:       data.StringMap{"key1": "value1", "key2": "value2"},
			expectError: false,
		},
		{
			name:        "map with special characters",
			input:       data.StringMap{"key with spaces": "value with spaces", "key\"quotes\"": "value\"quotes\""},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.input.Value()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.input == nil {
					assert.Nil(t, value)
				} else {
					// Verify it's valid JSON
					var parsed map[string]string
					jsonBytes, ok := value.([]byte)
					require.True(t, ok)
					err := json.Unmarshal(jsonBytes, &parsed)
					assert.NoError(t, err)
					assert.Equal(t, map[string]string(tt.input), parsed)
				}
			}
		})
	}
}

// Test StringArray and StringMap roundtrip (Scan -> Value -> Scan)
func TestStringArray_Roundtrip(t *testing.T) {
	original := data.StringArray{"item1", "item2", "item3"}

	// Convert to value
	value, err := original.Value()
	require.NoError(t, err)

	// Scan back
	var scanned data.StringArray
	err = scanned.Scan(value)
	require.NoError(t, err)

	assert.Equal(t, original, scanned)
}

func TestStringMap_Roundtrip(t *testing.T) {
	original := data.StringMap{"key1": "value1", "key2": "value2"}

	// Convert to value
	value, err := original.Value()
	require.NoError(t, err)

	// Scan back
	var scanned data.StringMap
	err = scanned.Scan(value)
	require.NoError(t, err)

	assert.Equal(t, original, scanned)
}

// Test model constants
func TestConstants(t *testing.T) {
	// Language constants
	assert.Equal(t, "en", data.LanguageEnglish)
	assert.Equal(t, "zh-TW", data.LanguageTraditionalChinese)

	// Interview type constants
	assert.Equal(t, "general", data.InterviewTypeGeneral)
	assert.Equal(t, "technical", data.InterviewTypeTechnical)
	assert.Equal(t, "behavioral", data.InterviewTypeBehavioral)
}

// Test edge cases for custom types
func TestStringArray_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{"empty byte slice", []byte{}},
		{"empty string", ""},
		{"whitespace string", "   "},
		{"null json", []byte("null")},
		{"null string", "null"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var arr data.StringArray
			err := arr.Scan(tt.input)
			// Should either succeed or fail gracefully
			if err != nil {
				assert.Error(t, err)
			} else {
				// If no error, result should be reasonable
				assert.NotPanics(t, func() { _, _ = arr.Value() })
			}
		})
	}
}

func TestStringMap_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{"empty byte slice", []byte{}},
		{"empty string", ""},
		{"whitespace string", "   "},
		{"null json", []byte("null")},
		{"null string", "null"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m data.StringMap
			err := m.Scan(tt.input)
			// Should either succeed or fail gracefully
			if err != nil {
				assert.Error(t, err)
			} else {
				// If no error, result should be reasonable
				assert.NotPanics(t, func() { _, _ = m.Value() })
			}
		})
	}
}

// Test that models are properly structured (basic structure validation)
func TestModelStructures(t *testing.T) { // Test Interview model
	interview := data.Interview{
		ID:                "test-id",
		CandidateName:     "John Doe",
		Questions:         data.StringArray{"Q1", "Q2"},
		InterviewLanguage: data.LanguageEnglish,
		Status:            "draft",
		InterviewType:     data.InterviewTypeGeneral,
		JobDescription:    "Software Engineer",
	}

	assert.Equal(t, "test-id", interview.ID)
	assert.Equal(t, "John Doe", interview.CandidateName)
	assert.Len(t, interview.Questions, 2)
	assert.Equal(t, data.LanguageEnglish, interview.InterviewLanguage)

	// Test Evaluation model
	evaluation := data.Evaluation{
		ID:          "eval-id",
		InterviewID: "interview-id",
		Answers:     data.StringMap{"Q1": "A1"},
		Score:       85.5,
		Feedback:    "Good performance",
	}

	assert.Equal(t, "eval-id", evaluation.ID)
	assert.Equal(t, 85.5, evaluation.Score)
	assert.Len(t, evaluation.Answers, 1)
	// Test ChatSession model
	session := data.ChatSession{
		ID:              "session-id",
		InterviewID:     "interview-id",
		SessionLanguage: data.LanguageEnglish,
		Status:          "active",
	}

	assert.Equal(t, "session-id", session.ID)
	assert.Equal(t, data.LanguageEnglish, session.SessionLanguage)

	// Test ChatMessage model
	message := data.ChatMessage{
		ID:        "msg-id",
		SessionID: "session-id",
		Type:      "user",
		Content:   "Hello",
	}

	assert.Equal(t, "msg-id", message.ID)
	assert.Equal(t, "user", message.Type)
	assert.Equal(t, "Hello", message.Content)
}

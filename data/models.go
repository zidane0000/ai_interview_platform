// Data models (structs for DB tables)
package data

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Language constants for interview support
const (
	LanguageEnglish            = "en"
	LanguageTraditionalChinese = "zh-TW"
)

// Interview type constants
const (
	InterviewTypeGeneral    = "general"
	InterviewTypeTechnical  = "technical"
	InterviewTypeBehavioral = "behavioral"
)

// ValidateLanguage checks if the provided language code is supported
func ValidateLanguage(lang string) bool {
	return lang == LanguageEnglish || lang == LanguageTraditionalChinese
}

// GetDefaultLanguage returns the default language when none is specified
func GetDefaultLanguage() string {
	return LanguageEnglish
}

// GetValidatedLanguage returns a valid language, defaulting to English if invalid
func GetValidatedLanguage(lang string) string {
	if ValidateLanguage(lang) {
		return lang
	}
	return GetDefaultLanguage()
}

// ValidateInterviewType checks if the provided interview type is supported
func ValidateInterviewType(interviewType string) bool {
	return interviewType == InterviewTypeGeneral ||
		interviewType == InterviewTypeTechnical ||
		interviewType == InterviewTypeBehavioral
}

// GetDefaultInterviewType returns the default interview type when none is specified
func GetDefaultInterviewType() string {
	return InterviewTypeGeneral
}

// GetValidatedInterviewType returns a valid interview type, defaulting to general if invalid
func GetValidatedInterviewType(interviewType string) string {
	if ValidateInterviewType(interviewType) {
		return interviewType
	}
	return GetDefaultInterviewType()
}

// StringArray is a custom type for handling PostgreSQL arrays with GORM
type StringArray []string

// Scan implements the Scanner interface for database/sql
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return fmt.Errorf("cannot scan %T into StringArray", value)
	}
}

// Value implements the Valuer interface for database/sql
func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// StringMap is a custom type for handling JSON maps with GORM
type StringMap map[string]string

// Scan implements the Scanner interface for database/sql
func (s *StringMap) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return fmt.Errorf("cannot scan %T into StringMap", value)
	}
}

// Value implements the Valuer interface for database/sql
func (s StringMap) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Interview model with proper GORM tags
type Interview struct {
	ID                string      `gorm:"primaryKey;type:varchar(255)" json:"id"`
	CandidateName     string      `gorm:"type:varchar(255);not null" json:"candidate_name"`
	Questions         StringArray `gorm:"type:jsonb" json:"questions"`
	InterviewLanguage string      `gorm:"column:language;type:varchar(10);not null;default:'en'" json:"interview_language"` // Interview language: "en" or "zh-TW"
	Status            string      `gorm:"type:varchar(50);not null;default:'draft'" json:"status"`                          // "draft", "active", "completed"
	InterviewType     string      `gorm:"column:type;type:varchar(50);not null" json:"interview_type"`                      // "general", "technical", "behavioral"
	JobDescription    string      `gorm:"type:text" json:"job_description,omitempty"`                                       // Optional: Job description text
	// TODO: Resume file support will be added in future iteration
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// Evaluation model with proper GORM tags
type Evaluation struct {
	ID          string    `gorm:"primaryKey;type:varchar(255)" json:"id"`
	InterviewID string    `gorm:"type:varchar(255);not null;index" json:"interview_id"`
	Answers     StringMap `gorm:"type:jsonb" json:"answers"`
	Score       float64   `gorm:"type:decimal(5,2)" json:"score"`
	Feedback    string    `gorm:"type:text" json:"feedback"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// ChatSession model for conversational interviews with proper GORM tags
type ChatSession struct {
	ID              string     `gorm:"primaryKey;type:varchar(255)" json:"id"`
	InterviewID     string     `gorm:"type:varchar(255);not null;index" json:"interview_id"`
	SessionLanguage string     `gorm:"column:language;type:varchar(10);not null;default:'en'" json:"session_language"` // Session language: "en" or "zh-TW"
	Status          string     `gorm:"type:varchar(50);not null;default:'active'" json:"status"`                       // "active", "completed", "abandoned"
	StartedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"started_at"`                             // When session started
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	EndedAt         *time.Time `gorm:"type:timestamp" json:"ended_at,omitempty"`
}

// ChatMessage model with proper GORM tags
type ChatMessage struct {
	ID        string    `gorm:"primaryKey;type:varchar(255)" json:"id"`
	SessionID string    `gorm:"type:varchar(255);not null;index" json:"session_id"`
	Type      string    `gorm:"type:varchar(50);not null" json:"type"` // "user", "ai"
	Content   string    `gorm:"type:text;not null" json:"content"`
	Timestamp time.Time `gorm:"not null" json:"timestamp"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TODO: Implement File model for resume uploads
// type File struct {
//     ID           string    `db:"id" json:"id"`
//     OriginalName string    `db:"original_name" json:"original_name"`
//     FileName     string    `db:"file_name" json:"file_name"`
//     FilePath     string    `db:"file_path" json:"file_path"`
//     FileSize     int64     `db:"file_size" json:"file_size"`
//     ContentType  string    `db:"content_type" json:"content_type"`
//     InterviewID  *string   `db:"interview_id" json:"interview_id,omitempty"`
//     CreatedAt    time.Time `db:"created_at" json:"created_at"`
// }

// TODO: Add database migration scripts
// TODO: Add indexes for performance optimization
// TODO: Add foreign key constraints
// TODO: Add validation tags for input validation
// TODO: Consider soft delete functionality (deleted_at fields)
// TODO: Add audit trail fields (created_by, updated_by)
// TODO: Add support for database transactions
// TODO: Add model conversion methods (ToDTO, FromDTO)

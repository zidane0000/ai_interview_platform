// Interview data access (CRUD operations)
package data

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// InterviewFilters defines filter options for interview queries
type InterviewFilters struct {
	CandidateName string
	Status        string
	Type          string
	CreatedAfter  time.Time
	CreatedBefore time.Time
}

// InterviewRepository interface defines the contract for interview data access
type InterviewRepository interface {
	Create(interview *Interview) error
	GetByID(id string) (*Interview, error)
	List(limit, offset int, filters InterviewFilters) ([]*Interview, int64, error)
	Update(id string, updates map[string]interface{}) error
	Delete(id string) error
	GetWithEvaluation(id string) (*Interview, *Evaluation, error)
}

// interviewRepository implements InterviewRepository interface
type interviewRepository struct {
	db *gorm.DB
}

// NewInterviewRepository creates a new interview repository
func NewInterviewRepository(db *gorm.DB) InterviewRepository {
	return &interviewRepository{db: db}
}

// Create creates a new interview
func (r *interviewRepository) Create(interview *Interview) error {
	interview.CreatedAt = time.Now()
	interview.UpdatedAt = time.Now()
	return r.db.Create(interview).Error
}

// GetByID retrieves an interview by ID
func (r *interviewRepository) GetByID(id string) (*Interview, error) {
	var interview Interview
	err := r.db.Where("id = ?", id).First(&interview).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("interview not found")
	}
	return &interview, err
}

// List retrieves interviews with pagination and filtering
func (r *interviewRepository) List(limit, offset int, filters InterviewFilters) ([]*Interview, int64, error) {
	var interviews []*Interview
	var total int64

	query := r.db.Model(&Interview{})
	// Apply filters
	if filters.CandidateName != "" {
		query = query.Where("candidate_name ILIKE ?", "%"+filters.CandidateName+"%")
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Type != "" {
		query = query.Where("type = ?", filters.Type)
	}
	if !filters.CreatedAfter.IsZero() {
		query = query.Where("created_at >= ?", filters.CreatedAfter)
	}
	if !filters.CreatedBefore.IsZero() {
		query = query.Where("created_at <= ?", filters.CreatedBefore)
	}

	// Get total count
	query.Count(&total)

	// Apply pagination and ordering
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&interviews).Error
	return interviews, total, err
}

// Update updates an interview
func (r *interviewRepository) Update(id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.Model(&Interview{}).Where("id = ?", id).Updates(updates).Error
}

// Delete deletes an interview (soft delete could be implemented here)
func (r *interviewRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&Interview{}).Error
}

// GetWithEvaluation retrieves an interview with its evaluation
func (r *interviewRepository) GetWithEvaluation(id string) (*Interview, *Evaluation, error) {
	var interview Interview
	var evaluation Evaluation

	err := r.db.Where("id = ?", id).First(&interview).Error
	if err != nil {
		return nil, nil, err
	}

	err = r.db.Where("interview_id = ?", id).First(&evaluation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &interview, nil, nil // No evaluation yet
	}

	return &interview, &evaluation, err
}

// TODO: Add database transaction support for complex operations
// TODO: Implement bulk operations (create, update, delete multiple records)
// TODO: Add database indexing recommendations in comments
// TODO: Implement audit logging for data changes
// TODO: Add caching layer for frequently accessed interviews
// TODO: Implement search functionality with full-text search
// TODO: Add data validation at repository level
// TODO: Implement data archival for old interviews

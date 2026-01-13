// Evaluation data access (CRUD operations)
package data

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// EvaluationFilters defines filter options for evaluation queries
type EvaluationFilters struct {
	InterviewID   string
	MinScore      float64
	MaxScore      float64
	CreatedAfter  time.Time
	CreatedBefore time.Time
}

// EvaluationStatistics provides aggregated statistics for evaluations
type EvaluationStatistics struct {
	TotalEvaluations  int64          `json:"total_evaluations"`
	AverageScore      float64        `json:"average_score"`
	MinScore          float64        `json:"min_score"`
	MaxScore          float64        `json:"max_score"`
	ScoreDistribution map[string]int `json:"score_distribution"` // Score ranges
}

// EvaluationRepository interface defines the contract for evaluation data access
type EvaluationRepository interface {
	Create(evaluation *Evaluation) error
	GetByID(id string) (*Evaluation, error)
	GetByInterviewID(interviewID string) (*Evaluation, error)
	List(limit, offset int, filters EvaluationFilters) ([]*Evaluation, int64, error)
	Update(id string, updates map[string]interface{}) error
	Delete(id string) error
	GetStatistics() (*EvaluationStatistics, error)
}

// evaluationRepository implements EvaluationRepository interface
type evaluationRepository struct {
	db *gorm.DB
}

// NewEvaluationRepository creates a new evaluation repository
func NewEvaluationRepository(db *gorm.DB) EvaluationRepository {
	return &evaluationRepository{db: db}
}

// Create creates a new evaluation with validation
func (r *evaluationRepository) Create(evaluation *Evaluation) error {
	// Validate that interview exists
	var interview Interview
	if err := r.db.Where("id = ?", evaluation.InterviewID).First(&interview).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("interview not found")
		}
		return err
	}

	evaluation.CreatedAt = time.Now()
	evaluation.UpdatedAt = time.Now()

	return r.db.Create(evaluation).Error
}

// GetByID retrieves an evaluation by ID
func (r *evaluationRepository) GetByID(id string) (*Evaluation, error) {
	var evaluation Evaluation
	err := r.db.Where("id = ?", id).First(&evaluation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("evaluation not found")
	}
	return &evaluation, err
}

// GetByInterviewID retrieves an evaluation by interview ID for frontend requirements
func (r *evaluationRepository) GetByInterviewID(interviewID string) (*Evaluation, error) {
	var evaluation Evaluation
	err := r.db.Where("interview_id = ?", interviewID).First(&evaluation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("evaluation not found")
	}
	return &evaluation, err
}

// List retrieves evaluations with filtering and sorting
func (r *evaluationRepository) List(limit, offset int, filters EvaluationFilters) ([]*Evaluation, int64, error) {
	var evaluations []*Evaluation
	var total int64

	query := r.db.Model(&Evaluation{})

	// Apply filters
	if filters.InterviewID != "" {
		query = query.Where("interview_id = ?", filters.InterviewID)
	}
	if filters.MinScore > 0 {
		query = query.Where("score >= ?", filters.MinScore)
	}
	if filters.MaxScore > 0 {
		query = query.Where("score <= ?", filters.MaxScore)
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
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&evaluations).Error
	return evaluations, total, err
}

// Update updates an evaluation
func (r *evaluationRepository) Update(id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.Model(&Evaluation{}).Where("id = ?", id).Updates(updates).Error
}

// Delete deletes an evaluation
func (r *evaluationRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&Evaluation{}).Error
}

// GetStatistics implements statistics aggregation for analytics
func (r *evaluationRepository) GetStatistics() (*EvaluationStatistics, error) {
	var stats EvaluationStatistics

	// Total count
	r.db.Model(&Evaluation{}).Count(&stats.TotalEvaluations)

	if stats.TotalEvaluations == 0 {
		return &stats, nil
	}

	// Average, min, max scores
	var result struct {
		Avg float64
		Min float64
		Max float64
	}

	err := r.db.Model(&Evaluation{}).
		Select("AVG(score) as avg, MIN(score) as min, MAX(score) as max").
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	stats.AverageScore = result.Avg
	stats.MinScore = result.Min
	stats.MaxScore = result.Max

	// Score distribution (simplified ranges)
	stats.ScoreDistribution = make(map[string]int)

	var distributions []struct {
		Range string
		Count int
	}

	err = r.db.Model(&Evaluation{}).
		Select(`
			CASE 
				WHEN score >= 90 THEN '90-100'
				WHEN score >= 80 THEN '80-89'
				WHEN score >= 70 THEN '70-79'
				WHEN score >= 60 THEN '60-69'
				ELSE '0-59'
			END as range,
			COUNT(*) as count
		`).
		Group("range").
		Scan(&distributions).Error

	if err != nil {
		return nil, err
	}

	for _, dist := range distributions {
		stats.ScoreDistribution[dist.Range] = dist.Count
	}

	return &stats, nil
}

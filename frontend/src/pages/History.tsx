import React, { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { formatDate } from '../utils/dateFormat';
import { logger } from '../utils/logger';
import axios from 'axios';
import {
  Box,
  Typography,
  Button,
  Card,
  CardContent,
  CardActions,
  CircularProgress,
  Alert,
  Container,
  Paper,
  Avatar,
  Divider,
  Fab,
  useTheme,
  alpha,
  TextField,
  InputAdornment,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Pagination,
  IconButton
} from '@mui/material';
import { 
  Add as AddIcon, 
  PlayArrow as PlayIcon, 
  Assessment as AssessmentIcon,
  AutoAwesome as AutoAwesomeIcon,
  Search as SearchIcon,
  ArrowBack as ArrowBackIcon,
  History as HistoryIcon
} from '@mui/icons-material';
import { interviewApi } from '../services/api';
import type { Interview } from '../types';

const History: React.FC = () => {
  const { t, i18n } = useTranslation();
  const [interviews, setInterviews] = useState<Interview[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [totalCount, setTotalCount] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(6);
  const [searchTerm, setSearchTerm] = useState('');
  const [sortBy, setSortBy] = useState<'date' | 'name' | 'status'>('date');
  const [sortOrder] = useState<'asc' | 'desc'>('desc');
  const theme = useTheme();

  const fetchInterviews = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      
      const response = await interviewApi.getInterviews({
        page: currentPage,
        limit: pageSize,
        candidate_name: searchTerm.trim() || undefined,
        sort_by: sortBy,
        sort_order: sortOrder
      });
      
      setInterviews(response.interviews);
      setTotalCount(response.total);
      
      logger.componentDebug('History', 'interviews fetched', {
        count: response.interviews.length,
        total: response.total,
        page: currentPage
      });
    } catch (err) {
      setError('Failed to load interviews');
      logger.error('Error fetching interviews', {
        component: 'History',
        action: 'fetchInterviews',
        data: err
      });
      
      if (axios.isAxiosError(err) && err.code === 'ECONNREFUSED') {
        setError('Unable to connect to server. Please make sure the backend is running.');
      }
    } finally {
      setLoading(false);
    }
  }, [currentPage, pageSize, searchTerm, sortBy, sortOrder]);

  useEffect(() => {
    fetchInterviews();
  }, [fetchInterviews]);

  const handleSearch = (event: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(event.target.value);
    setCurrentPage(1); // Reset to first page when searching
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const handleSortChange = (event: any) => {
    setSortBy(event.target.value as 'date' | 'name' | 'status');
    setCurrentPage(1);
  };

  const handlePageChange = (_event: React.ChangeEvent<unknown>, page: number) => {
    setCurrentPage(page);
  };

  const totalPages = Math.ceil(totalCount / pageSize);

  const getInterviewTypeColor = (type: string) => {
    switch (type) {
      case 'general': return theme.palette.primary.main;
      case 'technical': return theme.palette.success.main;
      case 'behavioral': return theme.palette.warning.main;
      default: return theme.palette.primary.main;
    }
  };

  const getInterviewTypeIcon = (type: string) => {
    switch (type) {
      case 'general': return '💼';
      case 'technical': return '💻';
      case 'behavioral': return '🧠';
      default: return '💼';
    }
  };

  return (
    <Box sx={{ minHeight: '100vh', backgroundColor: 'background.default' }}>
      {/* Hero Section */}
      <Box
        sx={{
          background: `linear-gradient(135deg, ${theme.palette.primary.main} 0%, ${theme.palette.secondary.main} 100%)`,
          color: '#ffffff',
          py: 6,
          position: 'relative',
          overflow: 'hidden',
          '&::after': {
            content: '""',
            position: 'absolute',
            bottom: 0,
            left: 0,
            right: 0,
            height: '100px',
            background: 'linear-gradient(0deg, rgba(255,255,255,0.1) 0%, transparent 100%)',
            pointerEvents: 'none'
          }
        }}
      >
        <Container maxWidth="lg">
          <Box textAlign="center">
            <Box display="flex" alignItems="center" justifyContent="center" mb={2}>
              <IconButton component={Link} to="/" sx={{ mr: 2, color: '#ffffff' }}>
                <ArrowBackIcon />
              </IconButton>
              <HistoryIcon sx={{ fontSize: 64, opacity: 0.9 }} />
            </Box>
            <Typography variant="h3" component="h1" gutterBottom fontWeight="bold">
              {t('pages:history.title', { defaultValue: 'Interview History' })}
            </Typography>
            <Typography variant="h6" sx={{ opacity: 0.95, maxWidth: 600, mx: 'auto' }}>
              {t('pages:history.subtitle', { defaultValue: 'Review your past interviews and track your progress' })}
            </Typography>
          </Box>
        </Container>
      </Box>

      <Container maxWidth="lg" sx={{ py: 6 }}>
        {/* Search and Filter Section */}
        <Paper 
          elevation={0} 
          sx={{ 
            p: 3, 
            mb: 4, 
            borderRadius: 3,
            border: `1px solid ${alpha(theme.palette.grey[300], 0.5)}`,
            background: 'linear-gradient(135deg, rgba(255,255,255,0.9) 0%, rgba(248,250,252,0.9) 100%)'
          }}
        >
          <Box 
            sx={{
              display: 'grid',
              gridTemplateColumns: { xs: '1fr', md: '2fr 1fr' },
              gap: 2,
              alignItems: 'center'
            }}
          >
            <TextField
              fullWidth
              placeholder={t('pages:history.searchPlaceholder', { defaultValue: 'Search by candidate name...' })}
              value={searchTerm}
              onChange={handleSearch}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchIcon color="action" />
                  </InputAdornment>
                ),
              }}
              sx={{
                '& .MuiOutlinedInput-root': {
                  borderRadius: 2,
                  backgroundColor: 'rgba(255,255,255,0.8)',
                }
              }}
            />
            <FormControl fullWidth>
              <InputLabel>{t('pages:history.sortBy', { defaultValue: 'Sort by' })}</InputLabel>
              <Select
                value={sortBy}
                onChange={handleSortChange}
                label={t('pages:history.sortBy', { defaultValue: 'Sort by' })}
                sx={{
                  borderRadius: 2,
                  backgroundColor: 'rgba(255,255,255,0.8)',
                }}
              >
                <MenuItem value="date">{t('pages:history.sortOptions.date', { defaultValue: 'Date' })}</MenuItem>
                <MenuItem value="name">{t('pages:history.sortOptions.name', { defaultValue: 'Name' })}</MenuItem>
              </Select>
            </FormControl>
          </Box>
        </Paper>

        {/* Content */}
        {loading ? (
          <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
            <CircularProgress size={48} />
          </Box>
        ) : error ? (
          <Alert 
            severity="error" 
            sx={{ 
              borderRadius: 2,
              '& .MuiAlert-message': { fontSize: '1rem' }
            }}
          >
            {error}
          </Alert>
        ) : interviews.length === 0 ? (
          <Paper 
            elevation={0}
            sx={{ 
              p: 6, 
              textAlign: 'center',
              borderRadius: 3,
              border: `1px solid ${alpha(theme.palette.grey[300], 0.5)}`,
              background: 'linear-gradient(135deg, rgba(255,255,255,0.9) 0%, rgba(248,250,252,0.9) 100%)'
            }}
          >
            <AutoAwesomeIcon sx={{ fontSize: 64, color: 'text.secondary', mb: 2 }} />
            <Typography variant="h5" gutterBottom color="text.secondary">
              {t('pages:history.noInterviews', { defaultValue: 'No interviews found' })}
            </Typography>
            <Typography variant="body1" color="text.secondary" paragraph>
              {searchTerm 
                ? t('pages:history.noSearchResults', { defaultValue: 'Try adjusting your search terms' })
                : t('pages:history.getStarted', { defaultValue: 'Create your first interview to get started!' })
              }
            </Typography>
            <Button 
              component={Link}
              to="/"
              variant="contained" 
              size="large"
              startIcon={<AddIcon />}
              sx={{ 
                mt: 2,
                borderRadius: 2,
                px: 4
              }}
            >
              {t('pages:history.createInterview', { defaultValue: 'Create Interview' })}
            </Button>
          </Paper>
        ) : (
          <>
            {/* Results Summary */}
            <Box sx={{ mb: 3 }}>
              <Typography variant="body2" color="text.secondary">
                {t('pages:history.resultsCount', { 
                  count: totalCount,
                  page: currentPage,
                  totalPages,
                  defaultValue: `Showing ${interviews.length} of ${totalCount} interviews (Page ${currentPage} of ${totalPages})`
                })}
              </Typography>
            </Box>

            {/* Interview Grid */}
            <Box
              sx={{
                display: 'grid',
                gridTemplateColumns: {
                  xs: '1fr',
                  sm: 'repeat(2, 1fr)',
                  lg: 'repeat(3, 1fr)'
                },
                gap: 3,
                mb: 4
              }}
            >
              {interviews.map((interview) => (
                <Card
                  key={interview.id}
                  elevation={0}
                  sx={{
                    borderRadius: 3,
                    border: `1px solid ${alpha(theme.palette.grey[300], 0.5)}`,
                    transition: 'all 0.3s ease',
                    '&:hover': {
                      transform: 'translateY(-2px)',
                      boxShadow: `0 8px 25px ${alpha(theme.palette.grey[400], 0.2)}`,
                    },
                    background: 'linear-gradient(135deg, rgba(255,255,255,0.9) 0%, rgba(248,250,252,0.9) 100%)'
                  }}
                >
                  <CardContent sx={{ p: 3 }}>
                    <Box display="flex" alignItems="center" gap={2} mb={2}>
                      <Avatar
                        sx={{
                          bgcolor: getInterviewTypeColor(interview.interview_type),
                          width: 48,
                          height: 48,
                          fontSize: '1.5rem'
                        }}
                      >
                        {getInterviewTypeIcon(interview.interview_type)}
                      </Avatar>
                      <Box>
                        <Typography variant="h6" fontWeight="600" noWrap>
                          {interview.candidate_name}
                        </Typography>
                        <Typography variant="body2" color="text.secondary">
                          {formatDate(interview.created_at, i18n.language)}
                        </Typography>
                      </Box>
                    </Box>

                    <Divider sx={{ my: 2 }} />

                    <Typography variant="body2" color="text.secondary" paragraph>
                      <strong>{t('pages:history.interviewType', { defaultValue: 'Type' })}:</strong> {interview.interview_type}
                    </Typography>

                    <Typography variant="body2" color="text.secondary" paragraph>
                      <strong>{t('pages:history.questionsCount', { defaultValue: 'Questions' })}:</strong> {interview.questions.length}
                    </Typography>

                    {interview.job_description && (
                      <Typography variant="body2" color="text.secondary" paragraph sx={{ 
                        display: '-webkit-box',
                        WebkitLineClamp: 2,
                        WebkitBoxOrient: 'vertical',
                        overflow: 'hidden'
                      }}>
                        <strong>{t('pages:history.jobDescription', { defaultValue: 'Job' })}:</strong> {interview.job_description}
                      </Typography>
                    )}
                  </CardContent>

                  <CardActions sx={{ p: 3, pt: 0 }}>
                    <Button
                      component={Link}
                      to={`/interview/${interview.id}`}
                      variant="outlined"
                      size="small"
                      startIcon={<AssessmentIcon />}
                      sx={{ borderRadius: 2 }}
                    >
                      {t('pages:history.viewDetails', { defaultValue: 'View Details' })}
                    </Button>
                    <Button
                      component={Link}
                      to={`/take-interview/${interview.id}`}
                      variant="contained"
                      size="small"
                      startIcon={<PlayIcon />}
                      sx={{ 
                        borderRadius: 2,
                        ml: 1
                      }}
                    >
                      {t('pages:history.startInterview', { defaultValue: 'Start' })}
                    </Button>
                  </CardActions>
                </Card>
              ))}
            </Box>

            {/* Pagination */}
            {totalPages > 1 && (
              <Box display="flex" justifyContent="center" mt={4}>
                <Pagination
                  count={totalPages}
                  page={currentPage}
                  onChange={handlePageChange}
                  color="primary"
                  size="large"
                  sx={{
                    '& .MuiPaginationItem-root': {
                      borderRadius: 2,
                    }
                  }}
                />
              </Box>
            )}
          </>
        )}

        {/* Floating Action Button */}
        <Fab
          component={Link}
          to="/"
          color="primary"
          aria-label="create interview"
          sx={{
            position: 'fixed',
            bottom: 24,
            right: 24,
            background: `linear-gradient(135deg, ${theme.palette.primary.main} 0%, ${theme.palette.primary.dark} 100%)`,
            '&:hover': {
              background: `linear-gradient(135deg, ${theme.palette.primary.dark} 0%, ${theme.palette.primary.main} 100%)`,
            }
          }}
        >
          <AddIcon />
        </Fab>
      </Container>
    </Box>
  );
};

export default History;

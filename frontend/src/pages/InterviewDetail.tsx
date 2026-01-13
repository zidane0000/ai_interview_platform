import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { formatDate } from '../utils/dateFormat';
import { logger } from '../utils/logger';
import {
  Typography,
  Button,
  Box,
  Chip,
  IconButton,
  Alert,
  CircularProgress,
  Card,
  CardContent,
  Container,
  Avatar,
  useTheme,
  alpha,
  Stack,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
} from '@mui/material';
import {
  ArrowBack as ArrowBackIcon,
  PlayArrow as PlayIcon,
  Quiz as QuizIcon,
  Schedule as ScheduleIcon,
  Assignment as AssignmentIcon,
  CheckCircle as CheckCircleIcon,
  Info as InfoIcon,
  Language as LanguageIcon,
} from '@mui/icons-material';
import { interviewApi } from '../services/api';
import type { Interview } from '../types';

const InterviewDetail: React.FC = () => {
  const { t, i18n } = useTranslation();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const theme = useTheme();  const [interview, setInterview] = useState<Interview | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedLanguage, setSelectedLanguage] = useState<'en' | 'zh-TW'>('en');

  const loadInterview = useCallback(async (interviewId: string) => {
    try {
      setLoading(true);
      const data = await interviewApi.getInterview(interviewId);
      setInterview(data);    } catch (err) {
      setError(t('common:errors.loadInterview'));
      logger.error('Error loading interview', {
        component: 'InterviewDetail',
        action: 'loadInterview',
        data: err
      });
    } finally {
      setLoading(false);
    }
  }, [t]);
  useEffect(() => {
    if (id) {
      loadInterview(id);
    }
  }, [id, loadInterview]);
  useEffect(() => {
    if (interview) {
      setSelectedLanguage(interview.interview_language || 'en');
    }
  }, [interview]);

  const handleStartInterview = () => {
    navigate(`/take-interview/${interview?.id}?lang=${selectedLanguage}`);
  };
  if (loading) {
    return (
      <Box 
        display="flex" 
        justifyContent="center" 
        alignItems="center" 
        minHeight="60vh"
        sx={{
          background: `linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.1)} 0%, ${alpha(theme.palette.secondary.main, 0.05)} 100%)`
        }}
      >
        <Box textAlign="center">
          <CircularProgress size={60} thickness={4} />
          <Typography variant="h6" sx={{ mt: 2, color: 'text.secondary' }}>
            {t('common:status.loading')}
          </Typography>
        </Box>
      </Box>
    );
  }

  if (error || !interview) {
    return (
      <Box sx={{ minHeight: '100vh' }}>
        {/* Header with gradient background */}
        <Box
          sx={{
            background: `linear-gradient(135deg, ${theme.palette.primary.main} 0%, ${theme.palette.secondary.main} 100%)`,
            color: '#ffffff',
            py: 4,
            position: 'relative',
            overflow: 'hidden',
          }}
        >
          <Container maxWidth="lg">
            <Box display="flex" alignItems="center">
              <IconButton 
                onClick={() => navigate('/')} 
                sx={{ 
                  mr: 2, 
                  color: '#ffffff',
                  '&:hover': { backgroundColor: 'rgba(255,255,255,0.1)' }
                }}
              >
                <ArrowBackIcon />
              </IconButton>
              <Typography variant="h4" component="h1" fontWeight="bold">
                {t('pages:interviewDetail.title')}
              </Typography>
            </Box>
          </Container>
        </Box>
        
        <Container maxWidth="lg" sx={{ py: 4 }}>
          <Alert 
            severity="error" 
            sx={{ 
              borderRadius: 3,
              boxShadow: `0 4px 20px ${alpha(theme.palette.error.main, 0.2)}`
            }}
          >
            {error || t('pages:interviewDetail.interviewNotFound')}
          </Alert>
        </Container>
      </Box>
    );
  }
  return (
    <Box sx={{ minHeight: '100vh', backgroundColor: 'background.default' }}>
      {/* Hero Section with Gradient Background */}
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
          <Box sx={{ position: 'relative', zIndex: 2 }}>
            {/* Navigation Header */}
            <Box display="flex" alignItems="center" mb={4}>
              <IconButton 
                onClick={() => navigate('/')} 
                sx={{ 
                  mr: 2, 
                  color: '#ffffff',
                  '&:hover': { backgroundColor: 'rgba(255,255,255,0.1)' }
                }}
              >
                <ArrowBackIcon />
              </IconButton>
              <Typography variant="h4" component="h1" fontWeight="bold">
                {t('pages:interviewDetail.title')}
              </Typography>
            </Box>

            {/* Candidate Hero Card */}
            <Card 
              elevation={0}
              sx={{ 
                backgroundColor: 'rgba(255,255,255,0.95)',
                backdropFilter: 'blur(10px)',
                borderRadius: 4,
                p: 4,
                boxShadow: `0 20px 60px ${alpha('#000000', 0.15)}`
              }}
            >
              <Box display="flex" alignItems="center" gap={3}>
                <Avatar
                  sx={{ 
                    width: 80, 
                    height: 80,
                    background: `linear-gradient(135deg, ${theme.palette.primary.main} 0%, ${theme.palette.primary.dark} 100%)`,
                    fontSize: '2rem',
                    fontWeight: 'bold',
                    boxShadow: `0 8px 32px ${alpha(theme.palette.primary.main, 0.3)}`
                  }}
                >
                  {interview.candidate_name.charAt(0).toUpperCase()}
                </Avatar>
                <Box flex={1}>
                  <Typography variant="h3" sx={{ color: 'text.primary', fontWeight: 700, mb: 1 }}>
                    {interview.candidate_name}
                  </Typography>
                  <Box display="flex" alignItems="center" gap={1} mb={2}>
                    <ScheduleIcon sx={{ color: 'text.secondary', fontSize: 20 }} />
                    <Typography variant="body1" color="text.secondary">
                      {t('pages:interviewDetail.createdAt')} {formatDate(interview.created_at, i18n.language)}
                    </Typography>
                  </Box>
                  
                  {/* Quick Stats */}
                  <Stack direction="row" spacing={2}>
                    <Chip
                      icon={<QuizIcon />}
                      label={`${interview.questions.length} ${t('pages:interviewDetail.questions')}`}
                      color="primary"
                      variant="outlined"
                      sx={{ 
                        borderRadius: 3,
                        backgroundColor: alpha(theme.palette.primary.main, 0.1),
                        fontWeight: 500
                      }}
                    />
                    <Chip
                      icon={<AssignmentIcon />}
                      label={t('pages:interviewDetail.interviewSummary')}
                      variant="outlined"
                      sx={{ 
                        borderRadius: 3,
                        backgroundColor: alpha(theme.palette.grey[500], 0.1),
                        fontWeight: 500
                      }}
                    />
                  </Stack>
                </Box>
                
                {/* Language Selection and Start Interview */}
                <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: 2 }}>
                  {/* Language Selection */}
                  <FormControl size="small" sx={{ minWidth: 200 }}>
                    <InputLabel sx={{ color: 'text.secondary' }}>
                      {t('pages:interviewDetail.interviewLanguage')}
                    </InputLabel>
                    <Select
                      value={selectedLanguage}
                      onChange={(e) => setSelectedLanguage(e.target.value as 'en' | 'zh-TW')}
                      label={t('pages:interviewDetail.interviewLanguage')}
                      startAdornment={<LanguageIcon sx={{ mr: 1, color: 'text.secondary' }} />}
                      sx={{
                        backgroundColor: 'background.paper',
                        borderRadius: 2,
                        '& .MuiOutlinedInput-notchedOutline': {
                          borderColor: alpha(theme.palette.grey[300], 0.5)
                        },
                        '&:hover .MuiOutlinedInput-notchedOutline': {
                          borderColor: theme.palette.primary.main
                        }
                      }}
                    >
                      <MenuItem value="en">
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                          <Box component="span" sx={{ fontSize: '1.2em' }}>🇺🇸</Box>
                          {t('common:languages.en')}
                        </Box>
                      </MenuItem>
                      <MenuItem value="zh-TW">
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                          <Box component="span" sx={{ fontSize: '1.2em' }}>🇹🇼</Box>
                          {t('common:languages.zh-TW')}
                        </Box>
                      </MenuItem>
                    </Select>
                  </FormControl>
                  
                  <Typography variant="caption" color="text.secondary" sx={{ textAlign: 'right', maxWidth: 200 }}>
                    {t('pages:interviewDetail.languageDescription')}
                  </Typography>

                  {/* Start Interview Button */}
                  <Button
                    variant="contained"
                    size="large"
                    startIcon={<PlayIcon />}
                    onClick={handleStartInterview}
                    sx={{
                      borderRadius: 4,
                      px: 4,
                      py: 2,
                      fontSize: '1.1rem',
                      fontWeight: 600,
                      background: `linear-gradient(135deg, ${theme.palette.success.main} 0%, ${theme.palette.success.dark} 100%)`,
                      boxShadow: `0 8px 32px ${alpha(theme.palette.success.main, 0.3)}`,
                      '&:hover': {
                        background: `linear-gradient(135deg, ${theme.palette.success.dark} 0%, ${theme.palette.success.main} 100%)`,
                        transform: 'translateY(-2px)',
                        boxShadow: `0 12px 40px ${alpha(theme.palette.success.main, 0.4)}`
                      },
                      transition: 'all 0.3s ease'
                    }}
                  >
                    {t('pages:interviewDetail.startInterview')}
                  </Button>
                </Box>
              </Box>
            </Card>
          </Box>
        </Container>
      </Box>

      {/* Main Content */}
      <Container maxWidth="lg" sx={{ py: 6 }}>
        <Box 
          sx={{
            display: 'grid',
            gridTemplateColumns: { xs: '1fr', lg: '2fr 1fr' },
            gap: 4,
          }}
        >
          {/* Questions Section */}
          <Box>
            <Card 
              elevation={0}
              sx={{ 
                borderRadius: 4,
                border: `1px solid ${alpha(theme.palette.grey[300], 0.5)}`,
                boxShadow: `0 4px 20px ${alpha(theme.palette.grey[500], 0.08)}`
              }}
            >
              <CardContent sx={{ p: 4 }}>
                <Box display="flex" alignItems="center" mb={3}>
                  <Box
                    sx={{
                      width: 48,
                      height: 48,
                      borderRadius: 3,
                      background: `linear-gradient(135deg, ${theme.palette.info.main} 0%, ${theme.palette.info.dark} 100%)`,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      mr: 2
                    }}
                  >
                    <QuizIcon sx={{ color: '#ffffff', fontSize: 24 }} />
                  </Box>
                  <Box>
                    <Typography variant="h5" fontWeight="700" gutterBottom>
                      {t('pages:interviewDetail.questions')}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      {t('pages:interviewDetail.questionsDescription')}
                    </Typography>
                  </Box>
                </Box>
                
                <Stack spacing={2}>
                  {interview.questions.map((question, index) => (
                    <Card 
                      key={index} 
                      variant="outlined" 
                      sx={{ 
                        borderRadius: 3,
                        border: `1px solid ${alpha(theme.palette.grey[300], 0.5)}`,
                        transition: 'all 0.2s ease',
                        '&:hover': {
                          boxShadow: `0 4px 20px ${alpha(theme.palette.grey[500], 0.1)}`,
                          borderColor: theme.palette.primary.main
                        }
                      }}
                    >
                      <CardContent sx={{ p: 3 }}>
                        <Box display="flex" alignItems="flex-start" gap={2}>
                          <Chip
                            label={`Q${index + 1}`}
                            size="small"
                            sx={{
                              backgroundColor: theme.palette.primary.main,
                              color: '#ffffff',
                              fontWeight: 600,
                              minWidth: '48px',
                              mt: 0.5
                            }}
                          />
                          <Typography 
                            variant="body1" 
                            sx={{ 
                              flexGrow: 1,
                              lineHeight: 1.6,
                              fontSize: '1rem'
                            }}
                          >
                            {question}
                          </Typography>
                        </Box>
                      </CardContent>
                    </Card>
                  ))}
                </Stack>
              </CardContent>
            </Card>
          </Box>

          {/* Sidebar */}
          <Box>
            <Stack spacing={3}>
              {/* Interview Summary Card */}
              <Card 
                elevation={0}
                sx={{ 
                  borderRadius: 4,
                  border: `1px solid ${alpha(theme.palette.grey[300], 0.5)}`,
                  boxShadow: `0 4px 20px ${alpha(theme.palette.grey[500], 0.08)}`
                }}
              >
                <CardContent sx={{ p: 4 }}>
                  <Box display="flex" alignItems="center" mb={3}>
                    <Box
                      sx={{
                        width: 40,
                        height: 40,
                        borderRadius: 2,
                        background: `linear-gradient(135deg, ${theme.palette.warning.main} 0%, ${theme.palette.warning.dark} 100%)`,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        mr: 2
                      }}
                    >
                      <AssignmentIcon sx={{ color: '#ffffff', fontSize: 20 }} />
                    </Box>
                    <Typography variant="h6" fontWeight="600">
                      {t('pages:interviewDetail.interviewSummary')}
                    </Typography>
                  </Box>
                  
                  <Box sx={{ mb: 3 }}>
                    <Typography variant="body2" color="text.secondary" gutterBottom>
                      {t('pages:interviewDetail.interviewId')}
                    </Typography>
                    <Typography 
                      variant="body1" 
                      sx={{ 
                        fontFamily: 'monospace', 
                        fontSize: '0.9rem',
                        backgroundColor: alpha(theme.palette.grey[100], 0.8),
                        px: 2,
                        py: 1,
                        borderRadius: 2,
                        border: `1px solid ${alpha(theme.palette.grey[300], 0.5)}`
                      }}
                    >
                      {interview.id}
                    </Typography>
                  </Box>
                </CardContent>
              </Card>

              {/* Instructions Card */}
              <Card 
                elevation={0}
                sx={{ 
                  borderRadius: 4,
                  border: `1px solid ${alpha(theme.palette.grey[300], 0.5)}`,
                  boxShadow: `0 4px 20px ${alpha(theme.palette.grey[500], 0.08)}`
                }}
              >
                <CardContent sx={{ p: 4 }}>
                  <Box display="flex" alignItems="center" mb={3}>
                    <Box
                      sx={{
                        width: 40,
                        height: 40,
                        borderRadius: 2,
                        background: `linear-gradient(135deg, ${theme.palette.info.main} 0%, ${theme.palette.info.dark} 100%)`,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        mr: 2
                      }}
                    >
                      <InfoIcon sx={{ color: '#ffffff', fontSize: 20 }} />
                    </Box>
                    <Typography variant="h6" fontWeight="600">
                      {t('pages:interviewDetail.instructions')}
                    </Typography>
                  </Box>
                  
                  <Stack spacing={1.5}>
                    {(t('pages:interviewDetail.instructionsList', { returnObjects: true }) as string[]).map((instruction: string, index: number) => (
                      <Box key={index} display="flex" alignItems="flex-start" gap={1.5}>
                        <CheckCircleIcon 
                          sx={{ 
                            color: theme.palette.success.main, 
                            fontSize: 18,
                            mt: 0.2,
                            flexShrink: 0
                          }} 
                        />
                        <Typography variant="body2" color="text.secondary" sx={{ lineHeight: 1.5 }}>
                          {instruction}
                        </Typography>
                      </Box>
                    ))}
                  </Stack>
                </CardContent>
              </Card>
            </Stack>
          </Box>
        </Box>
      </Container>
    </Box>
  );
};

export default InterviewDetail;

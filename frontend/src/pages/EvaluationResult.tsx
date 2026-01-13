import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { formatDate } from '../utils/dateFormat';
import { logger } from '../utils/logger';
import {
  Typography,
  Button,
  Box,
  IconButton,
  Alert,
  CircularProgress,
  Card,
  CardContent,
  Chip,
  LinearProgress,
  Container,
  Avatar,
  useTheme,
  alpha,
  Stack,
} from '@mui/material';

import {
  ArrowBack as ArrowBackIcon,
  Home as HomeIcon,
  Assessment as AssessmentIcon,
  CheckCircle as CheckIcon,
  EmojiEvents as TrophyIcon,
  Psychology as BrainIcon,
  Info as InfoIcon,
  Star as StarIcon,
} from '@mui/icons-material';
import { interviewApi } from '../services/api';
import type { Evaluation, Interview } from '../types';

const EvaluationResult: React.FC = () => {
  const { t, i18n } = useTranslation();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const theme = useTheme();
  const [evaluation, setEvaluation] = useState<Evaluation | null>(null);
  const [interview, setInterview] = useState<Interview | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);useEffect(() => {
    if (id) {
      loadEvaluation(id);
    }
  }, [id]); // eslint-disable-line react-hooks/exhaustive-deps
  const loadEvaluation = async (evaluationId: string) => {
    try {
      setLoading(true);
      const evaluationData = await interviewApi.getEvaluation(evaluationId);
      setEvaluation(evaluationData);
      
      // Load interview details to get candidate name
      const interviewData = await interviewApi.getInterview(evaluationData.interview_id);
      setInterview(interviewData);    } catch (err) {
      setError(t('pages:evaluationResult.failedToLoad'));
      logger.error('Error loading evaluation', {
        component: 'EvaluationResult',
        action: 'loadEvaluation',
        data: err
      });
    } finally {
      setLoading(false);
    }
  };
  const formatDateLocal = (dateString: string) => {
    return formatDate(dateString, i18n.language);
  };

  const getScoreColor = (score: number) => {
    if (score >= 0.8) return 'success';
    if (score >= 0.6) return 'warning';
    return 'error';
  };
  const getScoreLabel = (score: number) => {
    if (score >= 0.9) return t('pages:evaluationResult.scoreLabels.excellent');
    if (score >= 0.8) return t('pages:evaluationResult.scoreLabels.veryGood');
    if (score >= 0.7) return t('pages:evaluationResult.scoreLabels.good');
    if (score >= 0.6) return t('pages:evaluationResult.scoreLabels.fair');
    return t('pages:evaluationResult.scoreLabels.needsImprovement');
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

  if (error || !evaluation) {
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
                {t('pages:evaluationResult.title')}
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
            {error || t('pages:evaluationResult.evaluationNotFound')}
          </Alert>
        </Container>
      </Box>
    );
  }
  const scorePercentage = Math.round(evaluation.score * 100);

  return (
    <Box sx={{ minHeight: '100vh', backgroundColor: 'background.default' }}>
      {/* Hero Section with Results */}
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
                {t('pages:evaluationResult.title')}
              </Typography>
            </Box>

            {/* Results Hero Card */}
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
              <Box display="flex" alignItems="center" gap={4}>
                {/* Score Circle */}
                <Box 
                  sx={{ 
                    position: 'relative',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center'
                  }}
                >
                  <Avatar
                    sx={{ 
                      width: 120, 
                      height: 120,
                      background: `linear-gradient(135deg, ${theme.palette.success.main} 0%, ${theme.palette.success.dark} 100%)`,
                      fontSize: '2.5rem',
                      fontWeight: 'bold',
                      boxShadow: `0 8px 32px ${alpha(theme.palette.success.main, 0.3)}`
                    }}
                  >
                    {scorePercentage}%
                  </Avatar>
                  <TrophyIcon 
                    sx={{ 
                      position: 'absolute', 
                      top: -8, 
                      right: -8, 
                      color: '#FFD700', 
                      fontSize: 32,
                      filter: 'drop-shadow(0 2px 4px rgba(0,0,0,0.3))'
                    }} 
                  />
                </Box>                <Box flex={1}>
                  <Typography variant="h3" sx={{ color: 'text.primary', fontWeight: 700, mb: 1 }}>
                    {getScoreLabel(evaluation.score)}
                  </Typography>
                  {interview && (
                    <Typography variant="h5" sx={{ color: 'text.primary', fontWeight: 600, mb: 1 }}>
                      {interview.candidate_name}
                    </Typography>
                  )}
                  <Typography variant="h6" color="text.secondary" sx={{ mb: 2 }}>
                    {t('pages:evaluationResult.completedOn')} {formatDateLocal(evaluation.created_at)}
                  </Typography>
                  
                  {/* Performance Bar */}
                  <Box sx={{ mb: 2 }}>
                    <LinearProgress
                      variant="determinate"
                      value={scorePercentage}
                      color={getScoreColor(evaluation.score)}
                      sx={{ 
                        height: 12, 
                        borderRadius: 6,
                        backgroundColor: alpha(theme.palette.grey[300], 0.3),
                        '& .MuiLinearProgress-bar': {
                          borderRadius: 6
                        }
                      }}
                    />
                  </Box>
                  
                  {/* Quick Stats */}
                  <Stack direction="row" spacing={2}>
                    <Chip
                      icon={<AssessmentIcon />}
                      label={`${scorePercentage}% Score`}
                      color={getScoreColor(evaluation.score)}
                      variant="outlined"
                      sx={{ 
                        borderRadius: 3,
                        backgroundColor: alpha(theme.palette.success.main, 0.1),
                        fontWeight: 500
                      }}
                    />
                    <Chip
                      icon={<CheckIcon />}
                      label={t('pages:evaluationResult.completed')}
                      color="success"
                      variant="outlined"
                      sx={{ 
                        borderRadius: 3,
                        backgroundColor: alpha(theme.palette.success.main, 0.1),
                        fontWeight: 500
                      }}
                    />
                  </Stack>
                </Box>
                
                {/* Back to Home Button */}
                <Button
                  variant="contained"
                  size="large"
                  startIcon={<HomeIcon />}
                  onClick={() => navigate('/')}
                  sx={{
                    borderRadius: 4,
                    px: 4,
                    py: 2,
                    fontSize: '1.1rem',
                    fontWeight: 600,
                    background: `linear-gradient(135deg, ${theme.palette.primary.main} 0%, ${theme.palette.primary.dark} 100%)`,
                    boxShadow: `0 8px 32px ${alpha(theme.palette.primary.main, 0.3)}`,
                    '&:hover': {
                      background: `linear-gradient(135deg, ${theme.palette.primary.dark} 0%, ${theme.palette.primary.main} 100%)`,
                      transform: 'translateY(-2px)',
                      boxShadow: `0 12px 40px ${alpha(theme.palette.primary.main, 0.4)}`
                    },
                    transition: 'all 0.3s ease'
                  }}
                >
                  {t('pages:evaluationResult.backToHome')}
                </Button>
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
          {/* Main Feedback Section */}
          <Box>
            {/* AI Feedback Card */}
            <Card 
              elevation={0}
              sx={{ 
                borderRadius: 4,
                border: `1px solid ${alpha(theme.palette.grey[300], 0.5)}`,
                boxShadow: `0 4px 20px ${alpha(theme.palette.grey[500], 0.08)}`,
                mb: 4
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
                    <BrainIcon sx={{ color: '#ffffff', fontSize: 24 }} />
                  </Box>
                  <Box>
                    <Typography variant="h5" fontWeight="700" gutterBottom>
                      {t('pages:evaluationResult.aiFeedback')}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Detailed analysis and recommendations
                    </Typography>
                  </Box>
                </Box>
                
                <Card 
                  variant="outlined" 
                  sx={{ 
                    borderRadius: 3,
                    backgroundColor: alpha(theme.palette.info.main, 0.02),
                    border: `1px solid ${alpha(theme.palette.info.main, 0.2)}`
                  }}
                >
                  <CardContent sx={{ p: 3 }}>
                    <Typography 
                      variant="body1" 
                      sx={{ 
                        lineHeight: 1.7, 
                        whiteSpace: 'pre-wrap',
                        fontSize: '1.1rem'
                      }}
                    >
                      {evaluation.feedback}
                    </Typography>
                  </CardContent>
                </Card>
              </CardContent>
            </Card>

            {/* Performance Breakdown */}
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
                      background: `linear-gradient(135deg, ${theme.palette.warning.main} 0%, ${theme.palette.warning.dark} 100%)`,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      mr: 2
                    }}
                  >
                    <AssessmentIcon sx={{ color: '#ffffff', fontSize: 24 }} />
                  </Box>
                  <Typography variant="h5" fontWeight="700">
                    {t('pages:evaluationResult.performanceBreakdown')}
                  </Typography>
                </Box>
                
                <Box 
                  sx={{
                    display: 'grid',
                    gridTemplateColumns: { xs: 'repeat(2, 1fr)', sm: 'repeat(4, 1fr)' },
                    gap: 3,
                  }}
                >
                  <Card 
                    variant="outlined" 
                    sx={{ 
                      textAlign: 'center', 
                      p: 3,
                      borderRadius: 3,
                      backgroundColor: alpha(theme.palette.primary.main, 0.05),
                      border: `1px solid ${alpha(theme.palette.primary.main, 0.2)}`
                    }}
                  >
                    <Typography variant="h3" color="primary.main" fontWeight="bold">
                      {scorePercentage}
                    </Typography>
                    <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                      {t('pages:evaluationResult.overallScore')}
                    </Typography>
                  </Card>
                  
                  <Card 
                    variant="outlined" 
                    sx={{ 
                      textAlign: 'center', 
                      p: 3,
                      borderRadius: 3,
                      backgroundColor: alpha(theme.palette.info.main, 0.05),
                      border: `1px solid ${alpha(theme.palette.info.main, 0.2)}`
                    }}
                  >
                    <Typography variant="h3" color="info.main" fontWeight="bold">
                      A+
                    </Typography>
                    <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                      {t('pages:evaluationResult.grade')}
                    </Typography>
                  </Card>
                  
                  <Card 
                    variant="outlined" 
                    sx={{ 
                      textAlign: 'center', 
                      p: 3,
                      borderRadius: 3,
                      backgroundColor: alpha(theme.palette.success.main, 0.05),
                      border: `1px solid ${alpha(theme.palette.success.main, 0.2)}`
                    }}
                  >
                    <CheckIcon sx={{ fontSize: 40, color: 'success.main' }} />
                    <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                      {t('pages:evaluationResult.completed')}
                    </Typography>
                  </Card>
                  
                  <Card 
                    variant="outlined" 
                    sx={{ 
                      textAlign: 'center', 
                      p: 3,
                      borderRadius: 3,
                      backgroundColor: alpha(theme.palette.warning.main, 0.05),
                      border: `1px solid ${alpha(theme.palette.warning.main, 0.2)}`
                    }}
                  >
                    <BrainIcon sx={{ fontSize: 40, color: 'warning.main' }} />
                    <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                      {t('pages:evaluationResult.evaluated')}
                    </Typography>
                  </Card>
                </Box>
              </CardContent>
            </Card>
          </Box>

          {/* Sidebar */}
          <Box>
            <Stack spacing={3}>
              {/* Evaluation Details Card */}
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
                      {t('pages:evaluationResult.evaluationDetails')}
                    </Typography>
                  </Box>
                  
                  <Stack spacing={2}>
                    <Box>
                      <Typography variant="body2" color="text.secondary" gutterBottom>
                        {t('pages:evaluationResult.evaluationId')}
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
                        {evaluation.id}
                      </Typography>
                    </Box>

                    <Box>
                      <Typography variant="body2" color="text.secondary" gutterBottom>
                        {t('pages:evaluationResult.interviewId')}
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
                        {evaluation.interview_id}
                      </Typography>
                    </Box>

                    <Box>
                      <Typography variant="body2" color="text.secondary" gutterBottom>
                        {t('pages:evaluationResult.score')}
                      </Typography>
                      <Chip 
                        label={`${scorePercentage}% - ${getScoreLabel(evaluation.score)}`}
                        color={getScoreColor(evaluation.score)}
                        sx={{ fontWeight: 600 }}
                      />
                    </Box>

                    <Box>
                      <Typography variant="body2" color="text.secondary" gutterBottom>
                        {t('pages:evaluationResult.evaluationDate')}
                      </Typography>
                      <Typography variant="body1">
                        {formatDateLocal(evaluation.created_at)}
                      </Typography>
                    </Box>
                  </Stack>
                </CardContent>
              </Card>

              {/* Next Steps Card */}
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
                        background: `linear-gradient(135deg, ${theme.palette.success.main} 0%, ${theme.palette.success.dark} 100%)`,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        mr: 2
                      }}
                    >
                      <StarIcon sx={{ color: '#ffffff', fontSize: 20 }} />
                    </Box>
                    <Typography variant="h6" fontWeight="600">
                      {t('pages:evaluationResult.nextSteps')}
                    </Typography>
                  </Box>
                  
                  <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                    {t('pages:evaluationResult.nextStepsDescription')}
                  </Typography>
                  
                  <Stack spacing={1.5}>
                    {(t('pages:evaluationResult.nextStepsList', { returnObjects: true }) as string[]).map((step: string, index: number) => (
                      <Box key={index} display="flex" alignItems="flex-start" gap={1.5}>
                        <CheckIcon 
                          sx={{ 
                            color: theme.palette.success.main, 
                            fontSize: 18,
                            mt: 0.2,
                            flexShrink: 0
                          }} 
                        />
                        <Typography variant="body2" color="text.secondary" sx={{ lineHeight: 1.5 }}>
                          {step}
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

export default EvaluationResult;

import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate, useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { logger } from '../utils/logger';
import {
  Typography,
  TextField,
  Button,
  Box,
  IconButton,
  Alert,
  CircularProgress,
  Card,
  CardContent,
  Avatar,
  Container,
  useTheme,
  alpha,
  Stack,
  Chip,
} from '@mui/material';
import {
  ArrowBack as ArrowBackIcon,
  Send as SendIcon,
  SmartToy as AIIcon,
  Person as PersonIcon,
  Assessment as AssessmentIcon,
  Schedule as ScheduleIcon,
  Psychology as BrainIcon,
  Language as LanguageIcon,
} from '@mui/icons-material';
import { interviewApi } from '../services/api';
import type { Interview, ChatInterviewSession } from '../types';
import ErrorDisplay from '../components/ErrorDisplay';
import { createAppError } from '../services/errorService';
import type { AppError } from '../types/errors';

const TakeInterview: React.FC = () => {
  const { t } = useTranslation();
  const { id } = useParams<{ id: string }>();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const theme = useTheme();
  const [interview, setInterview] = useState<Interview | null>(null);
  const [chatSession, setChatSession] = useState<ChatInterviewSession | null>(null);
  const [currentMessage, setCurrentMessage] = useState('');
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const [error, setError] = useState<AppError | null>(null);
  const [interviewLanguage, setInterviewLanguage] = useState<'en' | 'zh-TW'>('en');
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const initializingRef = useRef(false);  useEffect(() => {
    if (id && !initializingRef.current) {
      initializingRef.current = true;
      // Extract language from URL parameters
      const langParam = searchParams.get('lang') as 'en' | 'zh-TW' | null;
      if (langParam && (langParam === 'en' || langParam === 'zh-TW')) {
        setInterviewLanguage(langParam);
      }
      initializeInterview(id, langParam || 'en');
    }
  }, [id, searchParams]);

  useEffect(() => {
    scrollToBottom();
  }, [chatSession?.messages]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };
  const initializeInterview = async (interviewId: string, language: 'en' | 'zh-TW' = 'en') => {
    try {
      setLoading(true);
      setError(null);
      
      // Load interview details
      const interviewData = await interviewApi.getInterview(interviewId);
      setInterview(interviewData);
        // Start chat session with language parameter
      const session = await interviewApi.startChatSession(interviewId, { session_language: language });
      setChatSession(session);    } catch (err) {
      const appError = createAppError(err, {
        component: 'TakeInterview',
        action: 'initializeInterview',
        fallbackType: 'server'
      });
      setError(appError);
      
      logger.error('Error initializing interview', {
        component: 'TakeInterview',
        action: 'initializeInterview',
        data: err
      });
    } finally {
      setLoading(false);
    }
  };  const handleSendMessage = async () => {
    if (!currentMessage.trim() || !chatSession || sending) return;

    const messageToSend = currentMessage.trim();
    
    try {
      setSending(true);
      setError(null);

      // 立即顯示用戶訊息
      const userMessage = {
        id: `temp_msg_${Date.now()}`,
        type: 'user' as const,
        content: messageToSend,
        timestamp: new Date().toISOString()
      };

      setChatSession(prev => prev ? {
        ...prev,
        messages: [...prev.messages, userMessage]
      } : null);

      // 清空輸入框
      setCurrentMessage('');      // 發送訊息到 API 並等待 AI 回應
      const response = await interviewApi.sendMessage(chatSession.id, {
        message: messageToSend
      });

      // 添加 AI 回應到聊天記錄並更新會話狀態
      if (response.ai_response) {
        const aiResponse = response.ai_response;
        setChatSession(prev => prev ? {
          ...prev,
          messages: [...prev.messages, aiResponse],
          status: response.session_status === 'completed' ? 'completed' : prev.status
        } : null);
      } else {
        // 如果沒有 AI 回應，仍然要更新會話狀態
        setChatSession(prev => prev ? {
          ...prev,
          status: response.session_status === 'completed' ? 'completed' : prev.status
        } : null);
      }    } catch (err) {
      const appError = createAppError(err, {
        component: 'TakeInterview',
        action: 'sendMessage',
        fallbackType: 'network'
      });
      setError(appError);
      
      logger.error('Error sending message', {
        component: 'TakeInterview',
        action: 'handleSendMessage',
        data: err
      });
      // 如果發生錯誤，恢復輸入內容
      setCurrentMessage(messageToSend);
    } finally {
      setSending(false);
    }
  };

  const handleEndInterview = async () => {
    if (!chatSession) return;

    try {
      setLoading(true);
      const evaluation = await interviewApi.endChatSession(chatSession.id);
      navigate(`/evaluation/${evaluation.id}`);    } catch (err) {
      const appError = createAppError(err, {
        component: 'TakeInterview',
        action: 'endInterview',
        fallbackType: 'server'
      });
      setError(appError);
      
      logger.error('Error ending interview', {
        component: 'TakeInterview',
        action: 'handleEndInterview',
        data: err
      });
      setLoading(false);
    }
  };

  const handleKeyPress = (event: React.KeyboardEvent) => {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      handleSendMessage();
    }
  };  if (loading) {
    return (
      <Box 
        display="flex" 
        justifyContent="center" 
        alignItems="center" 
        minHeight="100vh"
        sx={{
          background: `linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.1)} 0%, ${alpha(theme.palette.secondary.main, 0.05)} 100%)`
        }}
      >
        <Container maxWidth="sm">
          <Card 
            elevation={0}
            sx={{ 
              textAlign: 'center',
              p: 6,
              borderRadius: 4,
              backgroundColor: 'rgba(255,255,255,0.95)',
              backdropFilter: 'blur(10px)',
              boxShadow: `0 20px 60px ${alpha('#000000', 0.1)}`
            }}
          >
            <CircularProgress size={60} thickness={4} sx={{ mb: 3 }} />
            <Typography variant="h6" sx={{ mb: 2, color: 'text.primary', fontWeight: 600 }}>
              {interview ? t('pages:takeInterview.startingSession') : t('pages:takeInterview.loadingInterview')}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Please wait while we prepare your interview session...
            </Typography>
          </Card>
        </Container>
      </Box>
    );
  }
  if (error || !interview || !chatSession) {
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
                {t('pages:takeInterview.title')}
              </Typography>
            </Box>
          </Container>
        </Box>
        
        <Container maxWidth="lg" sx={{ py: 4 }}>
          {error ? (
            <ErrorDisplay 
              error={error}
              title="Interview Session Error"
              action="loadData"
              onRetry={() => id && initializeInterview(id, interviewLanguage)}
              showRetry={true}
            />
          ) : (
            <Alert 
              severity="error" 
              sx={{ 
                borderRadius: 3,
                boxShadow: `0 4px 20px ${alpha(theme.palette.error.main, 0.2)}`
              }}
            >
              {t('pages:takeInterview.sessionError')}
            </Alert>
          )}
        </Container>
      </Box>
    );
  }

  const isSessionCompleted = chatSession.status === 'completed';
  return (
    <Box sx={{ minHeight: '100vh', display: 'flex', flexDirection: 'column', backgroundColor: 'background.default' }}>
      {/* Modern Header with Interview Info */}
      <Box
        sx={{
          background: `linear-gradient(135deg, ${theme.palette.primary.main} 0%, ${theme.palette.secondary.main} 100%)`,
          color: '#ffffff',
          py: 3,
          position: 'relative',
          overflow: 'hidden',
          flexShrink: 0,
          boxShadow: `0 4px 20px ${alpha('#000000', 0.15)}`
        }}
      >
        <Container maxWidth="lg">
          <Box display="flex" alignItems="center" justifyContent="space-between">
            <Box display="flex" alignItems="center">
              <IconButton 
                onClick={() => navigate(`/interview/${interview.id}`)} 
                sx={{ 
                  mr: 2, 
                  color: '#ffffff',
                  '&:hover': { backgroundColor: 'rgba(255,255,255,0.1)' }
                }}
              >
                <ArrowBackIcon />
              </IconButton>
              <Box>
                <Typography variant="h5" component="h1" fontWeight="700" sx={{ mb: 0.5 }}>
                  {t('pages:takeInterview.interviewTitle', { candidate: interview.candidate_name })}
                </Typography>
                <Stack direction="row" spacing={2} alignItems="center">
                  <Chip
                    icon={isSessionCompleted ? <AssessmentIcon /> : <BrainIcon />}
                    label={isSessionCompleted ? t('pages:takeInterview.completed') : t('pages:takeInterview.inProgress')}
                    color={isSessionCompleted ? "success" : "info"}
                    variant="outlined"
                    sx={{ 
                      backgroundColor: 'rgba(255,255,255,0.15)',
                      color: '#ffffff',
                      borderColor: 'rgba(255,255,255,0.3)',
                      fontWeight: 600,
                      '& .MuiChip-icon': { color: '#ffffff' }
                    }}
                  />
                  <Chip
                    icon={<LanguageIcon />}
                    label={interviewLanguage === 'en' ? t('common:languages.en') : t('common:languages.zh-TW')}
                    variant="outlined"
                    sx={{ 
                      backgroundColor: 'rgba(255,255,255,0.1)',
                      color: '#ffffff',
                      borderColor: 'rgba(255,255,255,0.2)',
                      fontSize: '0.75rem',
                      '& .MuiChip-icon': { color: '#ffffff' }
                    }}
                  />
                  <Chip
                    icon={<ScheduleIcon />}
                    label={new Date().toLocaleTimeString()}
                    variant="outlined"
                    sx={{ 
                      backgroundColor: 'rgba(255,255,255,0.1)',
                      color: '#ffffff',
                      borderColor: 'rgba(255,255,255,0.2)',
                      fontSize: '0.75rem',
                      '& .MuiChip-icon': { color: '#ffffff' }
                    }}
                  />
                </Stack>
              </Box>
            </Box>
            
            {isSessionCompleted && (
              <Button
                variant="contained"
                size="large"
                onClick={handleEndInterview}
                disabled={loading}
                sx={{
                  borderRadius: 4,
                  px: 4,
                  py: 1.5,
                  fontSize: '1rem',
                  fontWeight: 600,
                  backgroundColor: 'rgba(255,255,255,0.2)',
                  backdropFilter: 'blur(10px)',
                  color: '#ffffff',
                  border: '1px solid rgba(255,255,255,0.3)',
                  boxShadow: `0 8px 32px ${alpha('#000000', 0.2)}`,
                  '&:hover': {
                    backgroundColor: 'rgba(255,255,255,0.3)',
                    transform: 'translateY(-1px)',
                    boxShadow: `0 12px 40px ${alpha('#000000', 0.3)}`
                  },
                  '&:disabled': {
                    backgroundColor: 'rgba(255,255,255,0.1)',
                    color: 'rgba(255,255,255,0.7)'
                  },
                  transition: 'all 0.3s ease'
                }}
              >
                {loading ? t('pages:takeInterview.processing') : t('pages:takeInterview.viewResults')}
              </Button>
            )}
          </Box>
        </Container>
      </Box>

      {/* Chat Messages Area */}
      <Box 
        sx={{ 
          flexGrow: 1, 
          overflow: 'auto', 
          background: `linear-gradient(180deg, ${alpha(theme.palette.grey[50], 0.5)} 0%, ${alpha(theme.palette.grey[100], 0.3)} 100%)`,
          position: 'relative'
        }}
      >
        <Container maxWidth="lg" sx={{ py: 3, height: '100%' }}>
          <Box sx={{ maxWidth: '800px', mx: 'auto' }}>
            {chatSession.messages.map((message) => (
              <Box
                key={message.id}
                sx={{
                  display: 'flex',
                  justifyContent: message.type === 'user' ? 'flex-end' : 'flex-start',
                  mb: 3
                }}
              >
                <Box
                  sx={{
                    display: 'flex',
                    alignItems: 'flex-start',
                    maxWidth: '75%',
                    flexDirection: message.type === 'user' ? 'row-reverse' : 'row'
                  }}
                >
                  <Avatar
                    sx={{
                      bgcolor: message.type === 'user' 
                        ? `linear-gradient(135deg, ${theme.palette.primary.main} 0%, ${theme.palette.primary.dark} 100%)`
                        : `linear-gradient(135deg, ${theme.palette.secondary.main} 0%, ${theme.palette.secondary.dark} 100%)`,
                      mx: 2,
                      width: 40,
                      height: 40,
                      boxShadow: `0 4px 12px ${alpha(message.type === 'user' ? theme.palette.primary.main : theme.palette.secondary.main, 0.3)}`
                    }}
                  >
                    {message.type === 'user' ? <PersonIcon /> : <AIIcon />}
                  </Avatar>
                  
                  <Card
                    elevation={0}
                    sx={{
                      background: message.type === 'user' 
                        ? `linear-gradient(135deg, ${theme.palette.primary.main} 0%, ${theme.palette.primary.dark} 100%)`
                        : '#ffffff',
                      color: message.type === 'user' ? '#ffffff' : 'text.primary',
                      borderRadius: 4,
                      border: message.type === 'user' 
                        ? 'none' 
                        : `1px solid ${alpha(theme.palette.grey[300], 0.5)}`,
                      boxShadow: `0 4px 20px ${alpha(message.type === 'user' ? theme.palette.primary.main : '#000000', 0.1)}`,
                      position: 'relative',
                      '&::before': message.type === 'user' ? {
                        content: '""',
                        position: 'absolute',
                        top: 16,
                        right: -8,
                        width: 0,
                        height: 0,
                        borderLeft: `8px solid ${theme.palette.primary.main}`,
                        borderTop: '8px solid transparent',
                        borderBottom: '8px solid transparent'
                      } : {
                        content: '""',
                        position: 'absolute',
                        top: 16,
                        left: -8,
                        width: 0,
                        height: 0,
                        borderRight: '8px solid #ffffff',
                        borderTop: '8px solid transparent',
                        borderBottom: '8px solid transparent'
                      }
                    }}
                  >
                    <CardContent sx={{ p: 3, '&:last-child': { pb: 3 } }}>
                      <Typography 
                        variant="body1" 
                        sx={{ 
                          whiteSpace: 'pre-wrap',
                          lineHeight: 1.6,
                          fontSize: '1rem'
                        }}
                      >
                        {message.content}
                      </Typography>
                      <Typography 
                        variant="caption" 
                        sx={{ 
                          opacity: 0.8, 
                          mt: 1.5, 
                          display: 'block',
                          fontSize: '0.75rem',
                          fontWeight: 500
                        }}
                      >
                        {new Date(message.timestamp).toLocaleTimeString()}
                      </Typography>
                    </CardContent>
                  </Card>
                </Box>
              </Box>
            ))}
            
            {sending && (
              <Box sx={{ display: 'flex', justifyContent: 'flex-start', mb: 3 }}>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <Avatar 
                    sx={{ 
                      background: `linear-gradient(135deg, ${theme.palette.secondary.main} 0%, ${theme.palette.secondary.dark} 100%)`,
                      mx: 2, 
                      width: 40, 
                      height: 40,
                      boxShadow: `0 4px 12px ${alpha(theme.palette.secondary.main, 0.3)}`
                    }}
                  >
                    <AIIcon />
                  </Avatar>
                  <Card 
                    elevation={0}
                    sx={{ 
                      background: '#ffffff',
                      p: 3,
                      borderRadius: 4,
                      border: `1px solid ${alpha(theme.palette.grey[300], 0.5)}`,
                      boxShadow: `0 4px 20px ${alpha('#000000', 0.08)}`,
                      position: 'relative',
                      '&::before': {
                        content: '""',
                        position: 'absolute',
                        top: 16,
                        left: -8,
                        width: 0,
                        height: 0,
                        borderRight: '8px solid #ffffff',
                        borderTop: '8px solid transparent',
                        borderBottom: '8px solid transparent'
                      }
                    }}
                  >
                    <Box display="flex" alignItems="center" gap={2}>
                      <Typography variant="body2" color="text.secondary" fontWeight={500}>
                        {t('pages:takeInterview.aiTyping')}
                      </Typography>
                      <CircularProgress size={16} thickness={4} />
                    </Box>
                  </Card>
                </Box>
              </Box>
            )}
            
            <div ref={messagesEndRef} />
          </Box>
        </Container>
      </Box>

      {/* Modern Message Input Area */}
      {!isSessionCompleted && (
        <Box
          sx={{
            backgroundColor: '#ffffff',
            borderTop: `1px solid ${alpha(theme.palette.grey[300], 0.5)}`,
            boxShadow: `0 -4px 20px ${alpha('#000000', 0.08)}`,
            flexShrink: 0
          }}
        >
          <Container maxWidth="lg" sx={{ py: 3 }}>
            <Box sx={{ maxWidth: '800px', mx: 'auto' }}>
              {error && (
                <ErrorDisplay 
                  error={error}
                  title="Message Error"
                  action="sendMessage"
                  onRetry={() => handleSendMessage()}
                  showRetry={currentMessage.trim().length > 0}
                  compact={true}
                />
              )}
              
              <Stack spacing={2}>
                <Box display="flex" gap={2} alignItems="flex-end">
                  <TextField
                    fullWidth
                    multiline
                    maxRows={4}
                    placeholder={t('pages:takeInterview.responsePlaceholder')}
                    value={currentMessage}
                    onChange={(e) => setCurrentMessage(e.target.value)}
                    onKeyPress={handleKeyPress}
                    disabled={sending}
                    sx={{
                      '& .MuiOutlinedInput-root': {
                        borderRadius: 4,
                        backgroundColor: alpha(theme.palette.grey[50], 0.5),
                        '&:hover': {
                          backgroundColor: alpha(theme.palette.grey[100], 0.8),
                        },
                        '&.Mui-focused': {
                          backgroundColor: '#ffffff',
                          boxShadow: `0 0 0 2px ${alpha(theme.palette.primary.main, 0.2)}`
                        }
                      }
                    }}
                  />
                  <Button
                    variant="contained"
                    size="large"
                    endIcon={<SendIcon />}
                    onClick={handleSendMessage}
                    disabled={!currentMessage.trim() || sending}
                    sx={{
                      minWidth: '120px',
                      borderRadius: 4,
                      height: '56px',
                      flexShrink: 0,
                      background: `linear-gradient(135deg, ${theme.palette.primary.main} 0%, ${theme.palette.primary.dark} 100%)`,
                      fontWeight: 600,
                      fontSize: '1rem',
                      boxShadow: `0 4px 20px ${alpha(theme.palette.primary.main, 0.3)}`,
                      '&:hover': {
                        background: `linear-gradient(135deg, ${theme.palette.primary.dark} 0%, ${theme.palette.primary.main} 100%)`,
                        transform: 'translateY(-1px)',
                        boxShadow: `0 8px 30px ${alpha(theme.palette.primary.main, 0.4)}`
                      },
                      '&:disabled': {
                        background: alpha(theme.palette.grey[400], 0.5),
                        color: alpha(theme.palette.text.primary, 0.5)
                      },
                      transition: 'all 0.3s ease'
                    }}
                  >
                    {sending ? 'Sending...' : 'Send'}
                  </Button>
                </Box>
                
                <Box display="flex" justifyContent="space-between" alignItems="center">
                  <Button
                    variant="outlined"
                    color="error"
                    size="medium"
                    onClick={handleEndInterview}
                    disabled={loading}
                    sx={{ 
                      borderRadius: 3,
                      textTransform: 'none',
                      fontWeight: 600,
                      px: 3,
                      py: 1,
                      borderColor: alpha(theme.palette.error.main, 0.5),
                      backgroundColor: alpha(theme.palette.error.main, 0.05),
                      '&:hover': {
                        backgroundColor: alpha(theme.palette.error.main, 0.1),
                        borderColor: theme.palette.error.main,
                        transform: 'translateY(-1px)',
                        boxShadow: `0 4px 20px ${alpha(theme.palette.error.main, 0.2)}`
                      },
                      transition: 'all 0.3s ease'
                    }}
                  >
                    {loading ? t('pages:takeInterview.processing') : t('pages:takeInterview.evaluate')}
                  </Button>
                  
                  <Typography 
                    variant="caption" 
                    color="text.secondary"
                    sx={{ 
                      fontStyle: 'italic',
                      fontSize: '0.75rem'
                    }}
                  >
                    Press Enter to send, Shift+Enter for new line
                  </Typography>
                </Box>
              </Stack>
            </Box>
          </Container>
        </Box>
      )}
    </Box>
  );
};

export default TakeInterview;

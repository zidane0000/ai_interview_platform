import React, { useState } from 'react';
import { Alert, AlertTitle, Button, Box, Chip, Collapse, Typography } from '@mui/material';
import { 
  Refresh as RefreshIcon, 
  ExpandMore as ExpandMoreIcon,
  ExpandLess as ExpandLessIcon,
  WifiOff as NetworkIcon,
  Error as ErrorIcon,
  Warning as WarningIcon,
  AccessTime as TimeoutIcon
} from '@mui/icons-material';
import type { AppError } from '../types/errors';
import { getContextualMessage, logError } from '../services/errorService';

interface ErrorDisplayProps {
  error: AppError | string;
  title?: string;
  onRetry?: () => void;
  showRetry?: boolean;
  action?: 'createInterview' | 'sendMessage' | 'loadData';
  compact?: boolean;
}

const ErrorDisplay: React.FC<ErrorDisplayProps> = ({ 
  error, 
  title,
  onRetry,
  showRetry = true,
  action,
  compact = false
}) => {
  const [showDetails, setShowDetails] = useState(false);
  const isDevelopment = import.meta.env.DEV;
  
  // Convert string error to AppError (memoized to prevent useEffect dependency issues)
  const appError: AppError = React.useMemo(() => 
    typeof error === 'string' 
      ? {
          type: 'client',
          userMessage: error,
          recoverable: true,
          timestamp: new Date(),
        }
      : error,
    [error]
  );
  
  // Log the error for monitoring
  React.useEffect(() => {
    if (typeof error !== 'string') {
      logError(appError);
    }
  }, [error, appError]);
  
  // Get appropriate user message
  const userMessage = action ? getContextualMessage(appError, action) : appError.userMessage;
  
  // Get error icon based on type
  const getErrorIcon = () => {
    switch (appError.type) {
      case 'network':
        return <NetworkIcon fontSize="small" />;
      case 'timeout':
        return <TimeoutIcon fontSize="small" />;
      case 'validation':
        return <WarningIcon fontSize="small" />;
      default:
        return <ErrorIcon fontSize="small" />;
    }
  };
  
  // Get error type chip color
  const getChipColor = () => {
    switch (appError.type) {
      case 'network':
      case 'timeout':
        return 'warning' as const;
      case 'validation':
        return 'info' as const;
      default:
        return 'error' as const;
    }
  };
  
  const handleRetry = () => {
    if (onRetry) {
      // Increment retry count
      appError.retryCount = (appError.retryCount || 0) + 1;
      onRetry();
    }
  };

  return (
    <Alert 
      severity="error" 
      sx={{ 
        mb: compact ? 1 : 2,
        '& .MuiAlert-message': {
          width: '100%'
        }
      }}
    >
      <AlertTitle sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
        {getErrorIcon()}
        {title || 'Something went wrong'}
        {!compact && (
          <Chip 
            label={appError.type} 
            size="small" 
            color={getChipColor()}
            variant="outlined"
            sx={{ ml: 'auto' }}
          />
        )}
      </AlertTitle>
      
      {/* User-friendly message */}
      <Typography variant="body2" sx={{ mb: compact ? 1 : 2 }}>
        {userMessage}
      </Typography>
      
      {/* Retry button and details */}
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, flexWrap: 'wrap' }}>
        {showRetry && onRetry && appError.recoverable && (
          <Button
            variant="outlined"
            size="small"
            startIcon={<RefreshIcon />}
            onClick={handleRetry}
            sx={{ minWidth: 'auto' }}
          >
            Try Again
            {appError.retryCount && appError.retryCount > 0 && (
              <Chip 
                label={appError.retryCount} 
                size="small" 
                sx={{ ml: 1, height: 16, fontSize: '0.7rem' }}
              />
            )}
          </Button>
        )}
        
        {/* Show details toggle (development only) */}
        {isDevelopment && appError.devMessage && !compact && (
          <Button
            size="small"
            variant="text"
            endIcon={showDetails ? <ExpandLessIcon /> : <ExpandMoreIcon />}
            onClick={() => setShowDetails(!showDetails)}
            sx={{ 
              minWidth: 'auto',
              color: 'text.secondary',
              '&:hover': {
                backgroundColor: 'action.hover'
              }
            }}
          >
            Details
          </Button>
        )}
      </Box>
      
      {/* Development details (collapsible) */}
      {isDevelopment && appError.devMessage && !compact && (
        <Collapse in={showDetails}>
          <Box sx={{ mt: 2, p: 1, backgroundColor: 'action.hover', borderRadius: 1 }}>
            <Typography variant="caption" color="text.secondary" sx={{ fontFamily: 'monospace' }}>
              <strong>Error Code:</strong> {appError.code || 'N/A'}<br />
              <strong>Timestamp:</strong> {appError.timestamp?.toLocaleTimeString()}<br />
              <strong>Component:</strong> {appError.context?.component || 'Unknown'}<br />
              <strong>Action:</strong> {appError.context?.action || 'Unknown'}<br />
              <strong>Dev Message:</strong> {appError.devMessage}
            </Typography>
          </Box>
        </Collapse>
      )}
    </Alert>
  );
};

export default ErrorDisplay;

/**
 * Error Service
 * Provides industry-standard error handling with user-friendly messages
 * Follows Netflix/Airbnb pattern: friendly for users, detailed for developers
 */

import { logger } from '../utils/logger';
import type { AppError } from '../types/errors';

/**
 * User-friendly error messages (Production)
 * These are what users see - always helpful and encouraging
 */
export const FRIENDLY_MESSAGES = {
  network: "We're having trouble connecting. Please check your internet connection and try again.",
  validation: "Please check your input and try again.",
  server: "Something went wrong on our end. We're working to fix it - please try again in a moment.",
  client: "Something unexpected happened. Refreshing the page might help.",
  timeout: "The request is taking longer than expected. Please try again.",
  
  // Context-specific messages for critical pages
  createInterview: "Unable to create interview right now. Please try again in a moment.",
  sendMessage: "Message couldn't be sent. Please check your connection and try again.",
  loadData: "Having trouble loading data. Please refresh the page or try again later.",
} as const;

/**
 * Developer error messages (Development)
 * These include technical details for debugging
 */
export const DEV_MESSAGES = {
  network: "Network request failed. Check console for details.",
  validation: "Validation failed. Check request payload.",
  server: "Server error occurred. Check API response and logs.",
  client: "Client-side error. Check component state and props.",
  timeout: "Request timeout. Check network conditions and API performance.",
} as const;

/**
 * Creates a standardized AppError from various error types
 */
export function createAppError(
  error: unknown,
  context?: {
    component?: string;
    action?: string;
    fallbackType?: AppError['type'];
  }
): AppError {
  const isDevelopment = import.meta.env.DEV;
  const timestamp = new Date();
  
  // Handle network/fetch errors
  if (error instanceof TypeError && error.message.includes('fetch')) {
    return {
      type: 'network',
      code: 'NETWORK_ERROR',
      userMessage: FRIENDLY_MESSAGES.network,
      devMessage: isDevelopment ? DEV_MESSAGES.network : undefined,
      recoverable: true,
      timestamp,
      context,
    };
  }
  
  // Handle API response errors
  if (error && typeof error === 'object' && 'status' in error) {
    const status = error.status as number;
    
    if (status >= 400 && status < 500) {
      return {
        type: 'validation',
        code: `HTTP_${status}`,
        userMessage: status === 404 ? 
          "The requested information wasn't found." : 
          FRIENDLY_MESSAGES.validation,
        devMessage: isDevelopment ? `HTTP ${status}: ${JSON.stringify(error)}` : undefined,
        recoverable: status !== 404,
        timestamp,
        context,
      };
    }
    
    if (status >= 500) {
      return {
        type: 'server',
        code: `HTTP_${status}`,
        userMessage: FRIENDLY_MESSAGES.server,
        devMessage: isDevelopment ? `HTTP ${status}: ${JSON.stringify(error)}` : undefined,
        recoverable: true,
        timestamp,
        context,
      };
    }
  }
  
  // Handle timeout errors
  if (error instanceof Error && error.name === 'AbortError') {
    return {
      type: 'timeout',
      code: 'REQUEST_TIMEOUT',
      userMessage: FRIENDLY_MESSAGES.timeout,
      devMessage: isDevelopment ? 'Request was aborted due to timeout' : undefined,
      recoverable: true,
      timestamp,
      context,
    };
  }
  
  // Handle generic errors
  const errorMessage = error instanceof Error ? error.message : String(error);
  
  return {
    type: context?.fallbackType || 'client',
    code: 'UNKNOWN_ERROR',
    userMessage: FRIENDLY_MESSAGES.client,
    devMessage: isDevelopment ? `Unknown error: ${errorMessage}` : undefined,
    recoverable: true,
    timestamp,
    context,
  };
}

/**
 * Gets context-specific error message for critical pages
 */
export function getContextualMessage(
  error: AppError,
  action?: 'createInterview' | 'sendMessage' | 'loadData'
): string {
  if (action && FRIENDLY_MESSAGES[action]) {
    return FRIENDLY_MESSAGES[action];
  }
  
  return error.userMessage;
}

/**
 * Logs error for monitoring (development) and production tracking
 */
export function logError(error: AppError): void {
  const logData = {
    type: error.type,
    code: error.code,
    component: error.context?.component,
    action: error.context?.action,
    timestamp: error.timestamp,
    retryCount: error.retryCount,
  };
  
  if (import.meta.env.DEV) {
    logger.error(error.devMessage || error.userMessage, {
      component: error.context?.component || 'ErrorService',
      data: logData,
    });
  } else {
    // In production, log without sensitive details
    logger.error('Application error occurred', {
      component: 'ErrorService',
      data: logData,
    });
  }
}

/**
 * Determines if an error should be retried based on type and context
 */
export function shouldRetryError(error: AppError): boolean {
  // Never retry validation errors
  if (error.type === 'validation') {
    return false;
  }
  
  // Always retry network and timeout errors
  if (error.type === 'network' || error.type === 'timeout') {
    return true;
  }
  
  // Retry server errors unless it's a specific non-retryable code
  if (error.type === 'server') {
    const nonRetryableCodes = ['HTTP_501', 'HTTP_502', 'HTTP_503'];
    return !nonRetryableCodes.includes(error.code || '');
  }
  
  return false;
}

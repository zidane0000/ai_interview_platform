/**
 * Auto-Retry Service
 * Conservative retry strategy with exponential backoff
 * Follows production best practices
 */

import { logger } from '../utils/logger';
import { createAppError, logError } from './errorService';
import type { RetryConfig, AppError } from '../types/errors';

/**
 * Delay function for retry backoff
 */
function delay(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * Calculate next retry delay with exponential backoff
 * Conservative approach: longer delays for production stability
 */
function calculateDelay(attempt: number, config: RetryConfig): number {
  const exponentialDelay = config.baseDelay * Math.pow(2, attempt);
  return Math.min(exponentialDelay, config.maxDelay);
}

/**
 * Auto-retry wrapper with conservative configuration
 * 
 * @param fn Function to retry
 * @param config Retry configuration (defaults to conservative)
 * @param context Error context for logging
 * @returns Promise with the result or throws AppError
 */
export async function retryWithBackoff<T>(
  fn: () => Promise<T>,
  config: Partial<RetryConfig> = {},
  context?: {
    component?: string;
    action?: string;
    description?: string;
  }
): Promise<T> {
  const retryConfig: RetryConfig = {
    maxRetries: 2, // Conservative: only 2 retries
    baseDelay: 1500, // Start with 1.5s delay
    maxDelay: 6000, // Cap at 6s delay
    retryCondition: (error) => {
      const errorType = 'type' in error ? error.type : undefined;
      return errorType === 'network' || errorType === 'timeout';
    },
    ...config,
  };
  
  let lastError: AppError | undefined;
  
  for (let attempt = 0; attempt <= retryConfig.maxRetries; attempt++) {
    try {
      if (attempt > 0) {
        const delayMs = calculateDelay(attempt - 1, retryConfig);
        
        logger.debug(`Retrying ${context?.description || 'operation'} (attempt ${attempt}/${retryConfig.maxRetries}) after ${delayMs}ms`, {
          component: context?.component || 'RetryService',
          action: context?.action || 'retry',
        });
        
        await delay(delayMs);
      }
      
      const result = await fn();
      
      // Log successful retry if this wasn't the first attempt
      if (attempt > 0) {
        logger.info(`${context?.description || 'Operation'} succeeded after ${attempt} retries`, {
          component: context?.component || 'RetryService',
          action: context?.action || 'retry_success',
        });
      }
      
      return result;
      
    } catch (error) {
      const appError = createAppError(error, {
        component: context?.component,
        action: context?.action,
      });
      
      // Add retry count to the error
      appError.retryCount = attempt;
      
      lastError = appError;
      
      // Check if we should retry this error
      if (attempt < retryConfig.maxRetries && retryConfig.retryCondition?.(appError)) {
        logger.warn(`${context?.description || 'Operation'} failed, will retry`, {
          component: context?.component || 'RetryService',
          action: context?.action || 'retry_attempt',
          data: {
            attempt: attempt + 1,
            maxRetries: retryConfig.maxRetries,
            errorType: appError.type,
            errorCode: appError.code,
          },
        });
        continue;
      }
      
      // Don't retry - either max attempts reached or error not retryable
      logError(appError);
      break;
    }
  }
  
  // All retries exhausted
  if (lastError) {
    logger.error(`${context?.description || 'Operation'} failed after ${retryConfig.maxRetries} retries`, {
      component: context?.component || 'RetryService',
      action: context?.action || 'retry_exhausted',
      data: {
        errorType: lastError.type,
        errorCode: lastError.code,
        totalAttempts: retryConfig.maxRetries + 1,
      },
    });
    
    throw lastError;
  }
  
  // This shouldn't happen, but TypeScript needs it
  throw createAppError(new Error('Retry logic error'), context);
}

/**
 * Retry specifically for API calls
 * Pre-configured for common API scenarios
 */
export async function retryApiCall<T>(
  apiCall: () => Promise<T>,
  endpoint: string,
  method: string = 'GET'
): Promise<T> {
  return retryWithBackoff(
    apiCall,
    {
      maxRetries: 2,
      baseDelay: 1500,
      maxDelay: 6000,
      retryCondition: (error) => {
        const errorType = 'type' in error ? error.type : undefined;
        const errorStatus = 'status' in error ? error.status : undefined;
        
        // Retry network errors, timeouts, and 5xx server errors
        return errorType === 'network' || 
               errorType === 'timeout' ||
               (errorStatus !== undefined && errorStatus >= 500 && errorStatus < 600);
      },
    },
    {
      component: 'ApiService',
      action: `${method.toUpperCase()}_${endpoint}`,
      description: `${method.toUpperCase()} ${endpoint}`,
    }
  );
}

/**
 * Retry for critical user actions
 * More aggressive retry for user-initiated actions
 */
export async function retryCriticalAction<T>(
  action: () => Promise<T>,
  actionName: string,
  component: string
): Promise<T> {
  return retryWithBackoff(
    action,
    {
      maxRetries: 3, // More retries for critical actions
      baseDelay: 1000, // Shorter initial delay
      maxDelay: 5000,
    },
    {
      component,
      action: actionName,
      description: actionName,
    }
  );
}

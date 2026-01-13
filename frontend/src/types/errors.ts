/**
 * Application Error Types
 * Follows industry patterns for production-grade error handling
 */

export interface AppError {
  type: 'network' | 'validation' | 'server' | 'client' | 'timeout';
  code?: string;
  userMessage: string;
  devMessage?: string;
  recoverable: boolean;
  retryCount?: number;
  timestamp?: Date;
  context?: {
    component?: string;
    action?: string;
    data?: unknown;
  };
}

export interface RetryConfig {
  maxRetries: number;
  baseDelay: number;
  maxDelay: number;
  retryCondition?: (error: AppError | Error | { type?: string; status?: number }) => boolean;
}

export const DEFAULT_RETRY_CONFIG: RetryConfig = {
  maxRetries: 2, // Conservative: 2 retries
  baseDelay: 1500, // Start with 1.5s delay
  maxDelay: 6000, // Cap at 6s delay
  retryCondition: (error) => {
    // Only retry network errors and 5xx server errors
    const errorType = 'type' in error ? error.type : undefined;
    const errorStatus = 'status' in error ? error.status : undefined;
    
    return errorType === 'network' || 
           errorType === 'timeout' ||
           (errorStatus !== undefined && errorStatus >= 500 && errorStatus < 600);
  }
};


/**
 * Centralized logging utility that respects environment settings
 * In production, debug and info logs are suppressed
 */

type LogLevel = 'debug' | 'info' | 'warn' | 'error';

interface LogContext {
  component?: string;
  action?: string;
  data?: unknown;
}

class Logger {
  private isDevelopment = import.meta.env.DEV;
  private isProduction = import.meta.env.PROD;

  private formatMessage(level: LogLevel, message: string, context?: LogContext): string {
    const timestamp = new Date().toISOString();
    const prefix = `[${timestamp}] [${level.toUpperCase()}]`;
    
    if (context?.component) {
      return `${prefix} [${context.component}] ${message}`;
    }
    
    return `${prefix} ${message}`;
  }

  private shouldLog(level: LogLevel): boolean {
    // In production, only log warnings and errors
    if (this.isProduction) {
      return level === 'warn' || level === 'error';
    }
    
    // In development, log everything
    return true;
  }

  debug(message: string, context?: LogContext): void {
    if (this.shouldLog('debug')) {
      console.log(this.formatMessage('debug', message, context), context?.data || '');
    }
  }

  info(message: string, context?: LogContext): void {
    if (this.shouldLog('info')) {
      console.info(this.formatMessage('info', message, context), context?.data || '');
    }
  }

  warn(message: string, context?: LogContext): void {
    if (this.shouldLog('warn')) {
      console.warn(this.formatMessage('warn', message, context), context?.data || '');
    }
  }

  error(message: string, context?: LogContext): void {
    if (this.shouldLog('error')) {
      console.error(this.formatMessage('error', message, context), context?.data || '');
    }
  }
  // Special method for API debugging - only in development
  apiDebug(endpoint: string, method: string, data?: unknown): void {
    if (this.isDevelopment) {
      this.debug(`API ${method} ${endpoint}`, {
        component: 'API',
        action: method,
        data: data
      });
    }
  }

  // Special method for component debugging - only in development
  componentDebug(component: string, action: string, data?: unknown): void {
    if (this.isDevelopment) {
      this.debug(`${action}`, {
        component: component,
        action: action,
        data: data
      });
    }
  }
}

// Export singleton instance
export const logger = new Logger();

// Export for testing
export { Logger };

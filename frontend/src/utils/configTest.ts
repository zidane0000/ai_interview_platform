// Simple test script to verify environment configuration
import { logger } from './logger';

logger.info('🔧 Environment Configuration Test');
logger.info('================================');
logger.info('MODE: ' + import.meta.env.MODE);
logger.info('VITE_API_BASE_URL: ' + import.meta.env.VITE_API_BASE_URL);
logger.info('VITE_USE_MOCK_DATA: ' + import.meta.env.VITE_USE_MOCK_DATA);
logger.info('VITE_DEV_MODE: ' + import.meta.env.VITE_DEV_MODE);
logger.info('VITE_APP_TITLE: ' + import.meta.env.VITE_APP_TITLE);

// Test mock data flag
const useMockData = import.meta.env.VITE_USE_MOCK_DATA === 'true';
logger.info('📊 Using Mock Data: ' + useMockData);

if (useMockData) {
  logger.info('🎭 Mock mode is active - no backend required');
} else {
  logger.info('🌐 API mode is active - backend required at: ' + import.meta.env.VITE_API_BASE_URL);
}

export {};

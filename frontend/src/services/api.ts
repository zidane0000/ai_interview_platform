import axios from 'axios';
import type {
  Interview,
  CreateInterviewRequest,
  SubmitEvaluationRequest,
  Evaluation,
  ListInterviewsResponse,
  ChatInterviewSession,
  SendMessageRequest,
  SendMessageResponse,
  StartChatSessionRequest,
} from '../types';
import { mockApi } from './mockApi';
import { logger } from '../utils/logger';
import { retryApiCall } from './retryService';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
const USE_MOCK_DATA = import.meta.env.VITE_USE_MOCK_DATA === 'true';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 10000, // 10 seconds timeout
});

// Request interceptor
api.interceptors.request.use(  (config) => {
    // Add any auth headers or other request modifications here
    logger.apiDebug(config.url || 'unknown', config.method?.toUpperCase() || 'UNKNOWN');
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor with enhanced error handling
api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    logger.error('API Error', {
      component: 'API',
      data: {
        url: error.config?.url,
        method: error.config?.method,
        status: error.response?.status,
        message: error.message,
      }
    });
    
    return Promise.reject(error);
  }
);

// Real API functions with retry logic for critical operations
const realApi = {
  // Create a new interview (CRITICAL - with retry)
  createInterview: async (data: CreateInterviewRequest): Promise<Interview> => {
    return retryApiCall(
      async () => {
        const response = await api.post<Interview>('/interviews', data);
        return response.data;
      },
      '/interviews',
      'POST'
    );
  },

  // Get all interviews with optional query parameters for pagination, filtering, and sorting
  getInterviews: async (params?: {
    limit?: number;
    offset?: number;
    page?: number;
    candidate_name?: string;
    status?: string;
    date_from?: string;
    date_to?: string;
    sort_by?: 'date' | 'name' | 'status';
    sort_order?: 'asc' | 'desc';
  }): Promise<ListInterviewsResponse> => {
    return retryApiCall(
      async () => {
        const response = await api.get<ListInterviewsResponse>('/interviews', { params });
        return response.data;
      },
      '/interviews',
      'GET'
    );
  },

  // Get specific interview
  getInterview: async (id: string): Promise<Interview> => {
    const response = await api.get<Interview>(`/interviews/${id}`);
    return response.data;
  },

  // Submit evaluation
  submitEvaluation: async (data: SubmitEvaluationRequest): Promise<Evaluation> => {
    const response = await api.post<Evaluation>('/evaluation', data);
    return response.data;
  },

  // Get evaluation results
  getEvaluation: async (id: string): Promise<Evaluation> => {
    const response = await api.get<Evaluation>(`/evaluation/${id}`);
    return response.data;
  },

  // Chat-based interview functions (CRITICAL - with retry)
  startChatSession: async (interviewId: string, options?: StartChatSessionRequest): Promise<ChatInterviewSession> => {
    return retryApiCall(
      async () => {
        const response = await api.post<ChatInterviewSession>(`/interviews/${interviewId}/chat/start`, options || {});
        return response.data;
      },
      `/interviews/${interviewId}/chat/start`,
      'POST'
    );
  },

  // Send message (CRITICAL - with retry)
  sendMessage: async (sessionId: string, data: SendMessageRequest): Promise<SendMessageResponse> => {
    return retryApiCall(
      async () => {
        const response = await api.post<SendMessageResponse>(`/chat/${sessionId}/message`, data);
        return response.data;
      },
      `/chat/${sessionId}/message`,
      'POST'
    );
  },

  getChatSession: async (sessionId: string): Promise<ChatInterviewSession> => {
    const response = await api.get<ChatInterviewSession>(`/chat/${sessionId}`);
    return response.data;
  },

  endChatSession: async (sessionId: string): Promise<Evaluation> => {
    const response = await api.post<Evaluation>(`/chat/${sessionId}/end`);
    return response.data;
  },
};

// Export the appropriate API based on environment
export const interviewApi = USE_MOCK_DATA ? mockApi : realApi;

export default api;

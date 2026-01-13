// TypeScript interfaces based on backend API DTOs
export interface Interview {
  id: string;
  candidate_name: string;
  questions: string[];
  interview_type: string; // Required: "general", "technical", or "behavioral"
  interview_language?: 'en' | 'zh-TW';
  job_description?: string; // Optional: Job description text
  created_at: string;
}

export interface CreateInterviewRequest {
  candidate_name: string;
  questions: string[];
  interview_type: string; // Required: "general", "technical", or "behavioral"
  interview_language?: 'en' | 'zh-TW';
  job_description?: string; // Optional: Job description text
}

export interface SubmitEvaluationRequest {
  interview_id: string;
  answers: Record<string, string>;
}

export interface Evaluation {
  id: string;
  interview_id: string;
  answers: Record<string, string>;
  score: number;
  feedback: string;
  created_at: string;
}

export interface ApiError {
  error: string;
  details?: string;
}

export interface ListInterviewsResponse {
  interviews: Interview[];
  total: number;
}

// Chat-based interview types
export interface ChatMessage {
  id: string;
  type: 'ai' | 'user';
  content: string;
  timestamp: string;
}

export interface ChatInterviewSession {
  id: string;
  interview_id: string;
  session_language?: 'en' | 'zh-TW';
  messages: ChatMessage[];
  status: 'active' | 'completed';
  created_at: string;
}

export interface SendMessageRequest {
  message: string;
}

export interface StartChatSessionRequest {
  session_language?: 'en' | 'zh-TW';
}

export interface SendMessageResponse {
  message: ChatMessage;
  ai_response?: ChatMessage;
  session_status: string;
}

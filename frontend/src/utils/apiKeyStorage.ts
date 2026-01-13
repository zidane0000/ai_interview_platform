// API Key storage helpers for BYOK (Bring Your Own Key) pattern
// Keys are stored in browser LocalStorage only - never sent to our servers

const OPENAI_KEY_STORAGE = 'ai_interview_openai_key';
const GEMINI_KEY_STORAGE = 'ai_interview_gemini_key';
const OPENAI_BASE_URL_STORAGE = 'ai_interview_openai_base_url';
const SELECTED_PROVIDER_STORAGE = 'ai_interview_selected_provider';

export type AIProvider = 'openai' | 'gemini' | 'mock';

// OpenAI key management
export const getOpenAIKey = (): string | null => {
  return localStorage.getItem(OPENAI_KEY_STORAGE);
};

export const setOpenAIKey = (key: string): void => {
  localStorage.setItem(OPENAI_KEY_STORAGE, key);
};

export const clearOpenAIKey = (): void => {
  localStorage.removeItem(OPENAI_KEY_STORAGE);
};

// OpenAI custom base URL management (for OpenAI-compatible providers)
export const getOpenAIBaseURL = (): string | null => {
  return localStorage.getItem(OPENAI_BASE_URL_STORAGE);
};

export const setOpenAIBaseURL = (url: string): void => {
  localStorage.setItem(OPENAI_BASE_URL_STORAGE, url);
};

export const clearOpenAIBaseURL = (): void => {
  localStorage.removeItem(OPENAI_BASE_URL_STORAGE);
};

// Gemini key management
export const getGeminiKey = (): string | null => {
  return localStorage.getItem(GEMINI_KEY_STORAGE);
};

export const setGeminiKey = (key: string): void => {
  localStorage.setItem(GEMINI_KEY_STORAGE, key);
};

export const clearGeminiKey = (): void => {
  localStorage.removeItem(GEMINI_KEY_STORAGE);
};

// Provider selection
export const getSelectedProvider = (): AIProvider => {
  const provider = localStorage.getItem(SELECTED_PROVIDER_STORAGE);
  if (provider === 'openai' || provider === 'gemini') {
    return provider;
  }
  return 'mock'; // Default to mock
};

export const setSelectedProvider = (provider: AIProvider): void => {
  localStorage.setItem(SELECTED_PROVIDER_STORAGE, provider);
};

// Check if user has any keys configured
export const hasAnyKeys = (): boolean => {
  return getOpenAIKey() !== null || getGeminiKey() !== null;
};

// Clear all keys
export const clearAllKeys = (): void => {
  clearOpenAIKey();
  clearGeminiKey();
  clearOpenAIBaseURL();
  localStorage.removeItem(SELECTED_PROVIDER_STORAGE);
};

// Get active key based on selected provider
export const getActiveKey = (): string | null => {
  const provider = getSelectedProvider();
  if (provider === 'openai') {
    return getOpenAIKey();
  } else if (provider === 'gemini') {
    return getGeminiKey();
  }
  return null;
};

// Validate key format
// For standard OpenAI, check for sk- prefix
// For custom endpoints, accept any non-empty key (JWT, Bearer, etc.)
export const validateOpenAIKey = (key: string, customBaseURL?: string): boolean => {
  if (!key || key.length === 0) {
    return false;
  }

  // If using custom base URL, accept any non-empty key format
  // (Together.ai, Groq, custom endpoints often use JWT or different formats)
  if (customBaseURL && customBaseURL.trim() !== '') {
    return key.length > 10;
  }

  // For standard OpenAI, require sk- prefix
  return key.startsWith('sk-') && key.length > 20;
};

export const validateGeminiKey = (key: string): boolean => {
  return key.length > 10; // Gemini keys don't have specific prefix
};

import { useState, useEffect } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  Typography,
  Box,
  Alert,
  Tabs,
  Tab,
} from '@mui/material';
import { useNavigate } from 'react-router-dom';
import {
  hasAnyKeys,
  setOpenAIKey,
  setGeminiKey,
  setOpenAIBaseURL,
  setSelectedProvider,
  validateOpenAIKey,
  validateGeminiKey,
  type AIProvider,
} from '../utils/apiKeyStorage';

const PROMPT_SHOWN_KEY = 'ai_interview_prompt_shown';

function ApiKeyPrompt() {
  const [open, setOpen] = useState(false);
  const [tab, setTab] = useState<AIProvider>('openai');
  const [openaiKey, setOpenaiKeyState] = useState('');
  const [openaiBaseURL, setOpenaiBaseURLState] = useState('');
  const [geminiKey, setGeminiKeyState] = useState('');
  const [error, setError] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    // Show prompt if:
    // 1. User has no keys configured
    // 2. Haven't seen prompt before
    const hasKeys = hasAnyKeys();
    const promptShown = localStorage.getItem(PROMPT_SHOWN_KEY);

    if (!hasKeys && !promptShown) {
      setOpen(true);
    }
  }, []);

  const handleSaveKey = () => {
    setError('');

    if (tab === 'openai') {
      if (!openaiKey) {
        setError('Please enter your OpenAI API key');
        return;
      }
      // Pass custom base URL to validation (relaxed validation for custom endpoints)
      if (!validateOpenAIKey(openaiKey, openaiBaseURL)) {
        if (openaiBaseURL) {
          setError('Invalid API key format (minimum 10 characters)');
        } else {
          setError('Invalid OpenAI API key (should start with sk-)');
        }
        return;
      }
      setOpenAIKey(openaiKey);

      // Save custom base URL if provided
      if (openaiBaseURL) {
        setOpenAIBaseURL(openaiBaseURL);
      }

      setSelectedProvider('openai');
    } else if (tab === 'gemini') {
      if (!geminiKey) {
        setError('Please enter your Gemini API key');
        return;
      }
      if (!validateGeminiKey(geminiKey)) {
        setError('Invalid Gemini API key');
        return;
      }
      setGeminiKey(geminiKey);
      setSelectedProvider('gemini');
    }

    localStorage.setItem(PROMPT_SHOWN_KEY, 'true');
    setOpen(false);
  };

  const handleUseMock = () => {
    setSelectedProvider('mock');
    localStorage.setItem(PROMPT_SHOWN_KEY, 'true');
    setOpen(false);
  };

  const handleGoToSettings = () => {
    localStorage.setItem(PROMPT_SHOWN_KEY, 'true');
    setOpen(false);
    navigate('/settings');
  };

  return (
    <Dialog open={open} maxWidth="sm" fullWidth disableEscapeKeyDown>
      <DialogTitle>
        <Typography variant="h5" sx={{ fontWeight: 700 }}>
          Welcome to AI Interview Platform
        </Typography>
      </DialogTitle>
      <DialogContent>
        <Typography variant="body1" sx={{ mb: 3 }}>
          To conduct interviews with real AI, you'll need to provide your own API key.
          Don't have one? You can try our free demo mode with mock AI responses.
        </Typography>

        <Tabs value={tab} onChange={(_, newValue) => setTab(newValue)} sx={{ mb: 2 }}>
          <Tab label="OpenAI" value="openai" />
          <Tab label="Gemini" value="gemini" />
        </Tabs>

        {tab === 'openai' && (
          <Box>
            <TextField
              fullWidth
              type="password"
              label="OpenAI API Key"
              placeholder="sk-..."
              value={openaiKey}
              onChange={(e) => setOpenaiKeyState(e.target.value)}
              helperText="Get your key from platform.openai.com/api-keys"
              sx={{ mb: 2 }}
            />
            <TextField
              fullWidth
              type="text"
              label="Custom API Base URL (Optional)"
              placeholder="https://api.openai.com/v1"
              value={openaiBaseURL}
              onChange={(e) => setOpenaiBaseURLState(e.target.value)}
              helperText="Leave empty for OpenAI, or use Together.ai, Groq, etc."
              sx={{ mb: 2 }}
            />
          </Box>
        )}

        {tab === 'gemini' && (
          <Box>
            <TextField
              fullWidth
              type="password"
              label="Gemini API Key"
              placeholder="Your Gemini API key"
              value={geminiKey}
              onChange={(e) => setGeminiKeyState(e.target.value)}
              helperText="Get your key from aistudio.google.com/app/apikey"
              sx={{ mb: 2 }}
            />
          </Box>
        )}

        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}

        <Alert severity="info">
          <strong>Privacy:</strong> Your API key is stored locally in your browser only.
          It's sent directly to {tab === 'openai' ? 'OpenAI' : 'Google'} via HTTPS.
          We never store or see your keys.
        </Alert>
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 3, flexDirection: 'column', gap: 1 }}>
        <Button
          variant="contained"
          fullWidth
          onClick={handleSaveKey}
          disabled={!openaiKey && !geminiKey}
        >
          Save & Continue with Real AI
        </Button>
        <Button
          variant="outlined"
          fullWidth
          onClick={handleUseMock}
        >
          Try Demo Mode (Mock AI)
        </Button>
        <Button
          size="small"
          onClick={handleGoToSettings}
        >
          Go to Settings Later
        </Button>
      </DialogActions>
    </Dialog>
  );
}

export default ApiKeyPrompt;

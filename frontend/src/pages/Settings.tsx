import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Container,
  Paper,
  Typography,
  TextField,
  Button,
  Box,
  Alert,
  RadioGroup,
  FormControlLabel,
  Radio,
  Divider,
  IconButton,
} from '@mui/material';
import { ArrowBack, Save, Delete } from '@mui/icons-material';
import {
  getOpenAIKey,
  setOpenAIKey,
  getGeminiKey,
  setGeminiKey,
  getOpenAIBaseURL,
  setOpenAIBaseURL,
  getSelectedProvider,
  setSelectedProvider,
  clearAllKeys,
  validateOpenAIKey,
  validateGeminiKey,
  type AIProvider,
} from '../utils/apiKeyStorage';

function Settings() {
  const navigate = useNavigate();

  const [provider, setProvider] = useState<AIProvider>('mock');
  const [openaiKey, setOpenaiKeyState] = useState('');
  const [openaiBaseURL, setOpenaiBaseURLState] = useState('');
  const [geminiKey, setGeminiKeyState] = useState('');
  const [saved, setSaved] = useState(false);
  const [error, setError] = useState('');

  // Load saved keys on mount
  useEffect(() => {
    const savedProvider = getSelectedProvider();
    const savedOpenAI = getOpenAIKey() || '';
    const savedOpenAIBaseURL = getOpenAIBaseURL() || '';
    const savedGemini = getGeminiKey() || '';

    setProvider(savedProvider);
    setOpenaiKeyState(savedOpenAI);
    setOpenaiBaseURLState(savedOpenAIBaseURL);
    setGeminiKeyState(savedGemini);
  }, []);

  const handleSave = () => {
    setError('');
    setSaved(false);

    // Validate based on selected provider
    if (provider === 'openai') {
      if (!openaiKey) {
        setError('OpenAI API key is required');
        return;
      }
      // Pass custom base URL to validation (relaxed validation for custom endpoints)
      if (!validateOpenAIKey(openaiKey, openaiBaseURL)) {
        if (openaiBaseURL) {
          setError('Invalid API key format (minimum 10 characters)');
        } else {
          setError('Invalid OpenAI API key format (should start with sk-)');
        }
        return;
      }
      setOpenAIKey(openaiKey);

      // Save custom base URL if provided (optional)
      if (openaiBaseURL) {
        setOpenAIBaseURL(openaiBaseURL);
      }
    } else if (provider === 'gemini') {
      if (!geminiKey) {
        setError('Gemini API key is required');
        return;
      }
      if (!validateGeminiKey(geminiKey)) {
        setError('Invalid Gemini API key format');
        return;
      }
      setGeminiKey(geminiKey);
    }

    // Save provider selection
    setSelectedProvider(provider);
    setSaved(true);

    // Auto-hide success message after 3 seconds
    setTimeout(() => setSaved(false), 3000);
  };

  const handleClear = () => {
    if (window.confirm('Are you sure you want to clear all API keys? You will return to mock AI mode.')) {
      clearAllKeys();
      setProvider('mock');
      setOpenaiKeyState('');
      setOpenaiBaseURLState('');
      setGeminiKeyState('');
      setSaved(false);
      setError('');
    }
  };

  return (
    <Container maxWidth="md" sx={{ py: 4 }}>
      <Box sx={{ mb: 3, display: 'flex', alignItems: 'center', gap: 2 }}>
        <IconButton onClick={() => navigate(-1)} size="large">
          <ArrowBack />
        </IconButton>
        <Typography variant="h4" sx={{ fontWeight: 700 }}>
          API Settings
        </Typography>
      </Box>

      <Paper sx={{ p: 4 }}>
        <Typography variant="h6" gutterBottom>
          Bring Your Own Key (BYOK)
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
          Use your own AI provider API keys. Keys are stored locally in your browser only and never sent to our servers.
        </Typography>

        <Alert severity="info" sx={{ mb: 3 }}>
          <strong>Privacy:</strong> Your API keys are stored in your browser's LocalStorage only.
          They are sent directly to AI providers (OpenAI/Google) via encrypted HTTPS.
          Our backend never stores your keys.
        </Alert>

        {saved && (
          <Alert severity="success" sx={{ mb: 2 }}>
            Settings saved successfully!
          </Alert>
        )}

        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}

        <Typography variant="subtitle2" gutterBottom sx={{ mt: 3, mb: 1 }}>
          AI Provider
        </Typography>
        <RadioGroup
          value={provider}
          onChange={(e) => setProvider(e.target.value as AIProvider)}
        >
          <FormControlLabel
            value="mock"
            control={<Radio />}
            label="Mock AI (Free demo mode - canned responses)"
          />
          <FormControlLabel
            value="openai"
            control={<Radio />}
            label="OpenAI (GPT-4 - requires your API key)"
          />
          <FormControlLabel
            value="gemini"
            control={<Radio />}
            label="Google Gemini (requires your API key)"
          />
        </RadioGroup>

        <Divider sx={{ my: 3 }} />

        {provider === 'openai' && (
          <Box sx={{ mb: 3 }}>
            <Typography variant="subtitle2" gutterBottom>
              OpenAI API Key
            </Typography>
            <TextField
              fullWidth
              type="password"
              placeholder="sk-..."
              value={openaiKey}
              onChange={(e) => setOpenaiKeyState(e.target.value)}
              helperText="Get your API key from platform.openai.com/api-keys"
              sx={{ mb: 2 }}
            />

            <Typography variant="subtitle2" gutterBottom>
              Custom API Base URL (Optional)
            </Typography>
            <TextField
              fullWidth
              type="text"
              placeholder="https://api.openai.com/v1"
              value={openaiBaseURL}
              onChange={(e) => setOpenaiBaseURLState(e.target.value)}
              helperText="Leave empty for OpenAI, or use custom OpenAI-compatible endpoint (Together.ai, Groq, etc.)"
            />
          </Box>
        )}

        {provider === 'gemini' && (
          <Box sx={{ mb: 3 }}>
            <Typography variant="subtitle2" gutterBottom>
              Gemini API Key
            </Typography>
            <TextField
              fullWidth
              type="password"
              placeholder="Your Gemini API key"
              value={geminiKey}
              onChange={(e) => setGeminiKeyState(e.target.value)}
              helperText="Get your API key from aistudio.google.com/app/apikey"
            />
          </Box>
        )}

        <Box sx={{ display: 'flex', gap: 2, mt: 4 }}>
          <Button
            variant="contained"
            startIcon={<Save />}
            onClick={handleSave}
            fullWidth
          >
            Save Settings
          </Button>
          <Button
            variant="outlined"
            color="error"
            startIcon={<Delete />}
            onClick={handleClear}
          >
            Clear All Keys
          </Button>
        </Box>

        <Alert severity="warning" sx={{ mt: 3 }}>
          <strong>Cost Note:</strong> When using your own API keys, you will be charged directly by the AI provider based on usage.
          Mock AI mode is always free.
        </Alert>
      </Paper>
    </Container>
  );
}

export default Settings;

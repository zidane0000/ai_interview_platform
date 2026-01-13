import { BrowserRouter as Router, Routes, Route, Navigate, Link } from 'react-router-dom';
import { ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { Container, AppBar, Toolbar, Typography, Box, Button } from '@mui/material';
import { useTranslation } from 'react-i18next';
import Home from './pages/Home';
import History from './pages/History';
import InterviewDetail from './pages/InterviewDetail';
import TakeInterview from './pages/TakeInterview';
import EvaluationResult from './pages/EvaluationResult';
import Changelog from './pages/Changelog';
import Settings from './pages/Settings';
import I18nTestPage from './components/I18nTestPage';
import LanguageSwitcher from './components/LanguageSwitcher';
import ModeIndicator from './components/ModeIndicator';
import ApiKeyPrompt from './components/ApiKeyPrompt';
import { theme } from './theme';

function App() {
  const { t } = useTranslation();
  
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Router>
        <Box sx={{ minHeight: '100vh', backgroundColor: 'background.default' }}>
          <AppBar 
            position="static" 
            elevation={0}
            sx={{
              background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
              borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
              borderRadius: 0
            }}
          >
            <Container maxWidth="lg">
              <Toolbar sx={{ px: 0 }}>
                <Typography 
                  variant="h6" 
                  component={Link}
                  to="/"
                  sx={{ 
                    flexGrow: 1,
                    fontWeight: 700,
                    fontSize: '1.25rem',
                    display: 'flex',
                    alignItems: 'center',
                    gap: 1,
                    color: 'inherit',
                    textDecoration: 'none',
                    '&:hover': { 
                      opacity: 0.8 
                    }
                  }}
                >
                  <span role="img" aria-label="robot">🤖</span>
                  {t('common:appName')}
                </Typography>
                
                <Box sx={{ display: 'flex', gap: 2, alignItems: 'center' }}>
                  <Button 
                    component={Link}
                    to="/history"
                    color="inherit"
                    sx={{ 
                      fontWeight: 500,
                      '&:hover': { backgroundColor: 'rgba(255,255,255,0.1)' }
                    }}
                  >
                    {t('common:navigation.history')}
                  </Button>
                  <Button
                    component={Link}
                    to="/changelog"
                    color="inherit"
                    sx={{
                      fontWeight: 500,
                      '&:hover': { backgroundColor: 'rgba(255,255,255,0.1)' }
                    }}
                  >
                    {t('common:navigation.changelog')}
                  </Button>
                  <Button
                    component={Link}
                    to="/settings"
                    color="inherit"
                    sx={{
                      fontWeight: 500,
                      '&:hover': { backgroundColor: 'rgba(255,255,255,0.1)' }
                    }}
                  >
                    Settings
                  </Button>
                  <LanguageSwitcher />
                </Box>
              </Toolbar>
            </Container>
          </AppBar>
          
          <ModeIndicator />
          <ApiKeyPrompt />

          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/history" element={<History />} />
            <Route path="/settings" element={<Settings />} />
            <Route path="/changelog" element={<Changelog />} />
            <Route path="/i18n-test" element={<I18nTestPage />} />
            <Route path="/interview/:id" element={<InterviewDetail />} />
            <Route path="/take-interview/:id" element={<TakeInterview />} />
            <Route path="/evaluation/:id" element={<EvaluationResult />} />
            {/* Redirect old mock-interview route to home for backward compatibility */}
            <Route path="/mock-interview" element={<Navigate to="/" replace />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </Box>
      </Router>
    </ThemeProvider>
  );
}

export default App

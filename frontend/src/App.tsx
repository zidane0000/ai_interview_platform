import { BrowserRouter as Router, Routes, Route, Navigate, Link } from 'react-router-dom';
import { ThemeProvider, createTheme } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { Container, AppBar, Toolbar, Typography, Box, Button } from '@mui/material';
import { useTranslation } from 'react-i18next';
import Home from './pages/Home';
import History from './pages/History';
import InterviewDetail from './pages/InterviewDetail';
import TakeInterview from './pages/TakeInterview';
import EvaluationResult from './pages/EvaluationResult';
import Changelog from './pages/Changelog';
import I18nTestPage from './components/I18nTestPage';
import LanguageSwitcher from './components/LanguageSwitcher';
import ModeIndicator from './components/ModeIndicator';

const theme = createTheme({
  palette: {
    primary: {
      main: '#6366f1', // Modern indigo
      light: '#818cf8',
      dark: '#4f46e5',
    },
    secondary: {
      main: '#f59e0b', // Modern amber
      light: '#fbbf24',
      dark: '#d97706',
    },
    background: {
      default: '#fafafb',
      paper: '#ffffff',
    },
    text: {
      primary: '#111827',
      secondary: '#6b7280',
    },
    error: {
      main: '#ef4444',
    },
    success: {
      main: '#10b981',
    },
    warning: {
      main: '#f59e0b',
    },
  },
  typography: {
    fontFamily: '"Inter", "Roboto", "Helvetica", "Arial", sans-serif',
    h1: {
      fontWeight: 800,
      fontSize: '2.5rem',
      letterSpacing: '-0.025em',
    },
    h2: {
      fontWeight: 700,
      fontSize: '2rem',
      letterSpacing: '-0.025em',
    },
    h3: {
      fontWeight: 600,
      fontSize: '1.5rem',
    },
    h4: {
      fontWeight: 600,
      fontSize: '1.25rem',
    },
    h5: {
      fontWeight: 500,
      fontSize: '1.125rem',
    },
    h6: {
      fontWeight: 500,
      fontSize: '1rem',
    },
    body1: {
      fontSize: '1rem',
      lineHeight: 1.7,
    },
    body2: {
      fontSize: '0.875rem',
      lineHeight: 1.6,
    },
  },
  shape: {
    borderRadius: 12,
  },
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          textTransform: 'none',
          fontWeight: 600,
          borderRadius: 8,
          padding: '10px 20px',
          boxShadow: 'none',
          '&:hover': {
            boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)',
          },
        },
        containedPrimary: {
          background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
          '&:hover': {
            background: 'linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%)',
          },
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          borderRadius: 16,
          boxShadow: '0 1px 3px rgba(0, 0, 0, 0.1)',
          '&:hover': {
            boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)',
          },
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          borderRadius: 16,
        },
      },
    },
    MuiChip: {
      styleOverrides: {
        root: {
          borderRadius: 8,
          fontWeight: 500,
        },
      },
    },
  },
});

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
                  <LanguageSwitcher />
                </Box>
              </Toolbar>
            </Container>
          </AppBar>
          
          <ModeIndicator />
          
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/history" element={<History />} />
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

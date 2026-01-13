import React from 'react';
import { useTranslation } from 'react-i18next';
import {
  Box,
  Typography,
  Paper,
  Button,
  Card,
  CardContent,
  Divider,
  Container
} from '@mui/material';
import {
  Language as LanguageIcon,
  CheckCircle as CheckCircleIcon
} from '@mui/icons-material';

const I18nTestPage: React.FC = () => {
  const { t, i18n } = useTranslation();

  const testKeys = [
    'common:appName',
    'common:buttons.create',
    'common:buttons.start',
    'common:navigation.home',
    'pages:home.title',
    'pages:home.subtitle',
    'pages:mockInterview.title',
    'pages:mockInterview.candidateName',
    'interview:evaluation.criteria.clarity',
    'interview:questions.types.technical'
  ];

  return (
    <Container maxWidth="md" sx={{ py: 4 }}>      <Paper elevation={2} sx={{ p: 4, mb: 4 }}>
        <Box display="flex" alignItems="center" mb={3}>
          <LanguageIcon sx={{ mr: 2, fontSize: 32, color: 'primary.main' }} />
          <Typography variant="h4" component="h1" fontWeight="600">
            {t('pages:i18nTest.pageTitle')}
          </Typography>
        </Box>
        
        <Typography variant="body1" color="text.secondary" sx={{ mb: 3 }}>
          {t('pages:i18nTest.currentLanguage')} <strong>{i18n.language}</strong>
        </Typography>

        <Divider sx={{ my: 3 }} />

        <Typography variant="h6" gutterBottom>
          {t('pages:i18nTest.translationKeys')}
        </Typography>
        
        <Box sx={{ display: 'grid', gap: 2 }}>
          {testKeys.map((key) => (
            <Card key={key} variant="outlined">
              <CardContent sx={{ py: 2 }}>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      {key}
                    </Typography>
                    <Typography variant="body1" fontWeight="500">
                      {t(key)}
                    </Typography>
                  </Box>
                  <CheckCircleIcon color="success" />
                </Box>
              </CardContent>
            </Card>
          ))}
        </Box>        <Divider sx={{ my: 3 }} />

        <Typography variant="h6" gutterBottom>
          {t('pages:i18nTest.languageSwitch')}
        </Typography>
        
        <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
          <Button
            variant={i18n.language === 'en' ? 'contained' : 'outlined'}
            onClick={() => i18n.changeLanguage('en')}
          >
            {t('common:languages.en')}
          </Button>
          <Button
            variant={i18n.language === 'zh-TW' ? 'contained' : 'outlined'}
            onClick={() => i18n.changeLanguage('zh-TW')}
          >
            {t('common:languages.zh-TW')}
          </Button>
        </Box>
      </Paper>
    </Container>
  );
};

export default I18nTestPage;

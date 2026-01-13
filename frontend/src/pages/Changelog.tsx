import React from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  Box,
  Typography,
  Container,
  Card,
  CardContent,
  Chip,
  useTheme,
  alpha,
  Button,
  Divider
} from '@mui/material';
import { 
  Update as UpdateIcon,
  BugReport as BugReportIcon,
  NewReleases as NewReleasesIcon,
  ArrowBack as ArrowBackIcon
} from '@mui/icons-material';

interface ChangelogEntry {
  version: string;
  date: string;
  type: 'improvement' | 'fix' | 'new';
  changes: string[];
}

const Changelog: React.FC = () => {
  const { t } = useTranslation();
  const theme = useTheme();
  const changelogData: ChangelogEntry[] = [
    {
      version: 'v1.0',
      date: t('pages:changelog.v1_0.date'),
      type: 'new',
      changes: t('pages:changelog.v1_0.changes', { returnObjects: true }) as string[]
    }
  ];

  const getTypeConfig = (type: ChangelogEntry['type']) => {
    switch (type) {
      case 'improvement':
        return {
          icon: <UpdateIcon />,
          label: t('pages:changelog.typeLabels.improvement'),
          color: theme.palette.primary.main,
          bgColor: alpha(theme.palette.primary.main, 0.1)
        };
      case 'fix':
        return {
          icon: <BugReportIcon />,
          label: t('pages:changelog.typeLabels.fix'),
          color: theme.palette.warning.main,
          bgColor: alpha(theme.palette.warning.main, 0.1)
        };
      case 'new':
        return {
          icon: <NewReleasesIcon />,
          label: t('pages:changelog.typeLabels.new'),
          color: theme.palette.success.main,
          bgColor: alpha(theme.palette.success.main, 0.1)
        };
      default:
        return {
          icon: <UpdateIcon />,
          label: t('pages:changelog.typeLabels.update'),
          color: theme.palette.primary.main,
          bgColor: alpha(theme.palette.primary.main, 0.1)
        };
    }
  };

  return (    <Box sx={{ minHeight: '100vh', backgroundColor: 'background.default' }}>
      {/* Hero Section */}
      <Box
        sx={{
          background: `linear-gradient(135deg, ${theme.palette.primary.main} 0%, ${theme.palette.secondary.main} 100%)`,
          color: '#ffffff',
          py: 6,
          position: 'relative',
          overflow: 'hidden',
          '&::before': {
            content: '""',
            position: 'absolute',
            top: '20%',
            right: '-10%',
            width: '300px',
            height: '300px',
            borderRadius: '50%',
            background: 'rgba(255,255,255,0.05)',
            pointerEvents: 'none'
          }
        }}
      >
        <Container maxWidth="lg">          <Box textAlign="center">
            <UpdateIcon sx={{ fontSize: 64, mb: 2, opacity: 0.9 }} />
            <Typography variant="h3" component="h1" gutterBottom fontWeight="bold">
              {t('pages:changelog.title')}
            </Typography>
            <Typography variant="h6" sx={{ opacity: 0.95, maxWidth: 600, mx: 'auto' }}>
              {t('pages:changelog.subtitle')}
            </Typography>
          </Box>
        </Container>
      </Box>

      <Container maxWidth="md" sx={{ py: 6 }}>
        {/* Back Button */}        <Button
          component={Link}
          to="/"
          startIcon={<ArrowBackIcon />}
          sx={{ mb: 4, borderRadius: 2 }}
        >
          {t('pages:changelog.backToHome')}
        </Button>

        {/* Changelog Entries */}
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
          {changelogData.map((entry, index) => {
            const typeConfig = getTypeConfig(entry.type);
            
            return (
              <Card 
                key={`${entry.version}-${index}`}
                elevation={0}
                sx={{ 
                  borderRadius: 3,
                  border: `1px solid ${alpha(theme.palette.grey[300], 0.5)}`,
                  transition: 'all 0.3s ease',
                  '&:hover': {
                    transform: 'translateY(-2px)',
                    boxShadow: `0 8px 25px ${alpha(theme.palette.grey[400], 0.15)}`
                  }
                }}
              >
                <CardContent sx={{ p: 4 }}>
                  {/* Header */}
                  <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 3 }}>
                    <Box>
                      <Typography variant="h5" fontWeight="bold" gutterBottom>
                        {entry.version}
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        {entry.date}
                      </Typography>
                    </Box>
                    <Chip
                      icon={typeConfig.icon}
                      label={typeConfig.label}
                      sx={{
                        backgroundColor: typeConfig.bgColor,
                        color: typeConfig.color,
                        fontWeight: 600,
                        '& .MuiChip-icon': {
                          color: typeConfig.color
                        }
                      }}
                    />
                  </Box>

                  <Divider sx={{ mb: 3 }} />

                  {/* Changes List */}
                  <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5 }}>
                    {entry.changes.map((change, changeIndex) => (
                      <Box 
                        key={changeIndex}
                        sx={{ 
                          display: 'flex', 
                          alignItems: 'flex-start', 
                          gap: 2,
                          p: 2,
                          borderRadius: 2,
                          backgroundColor: alpha(theme.palette.grey[50], 0.5),
                          border: `1px solid ${alpha(theme.palette.grey[200], 0.5)}`
                        }}
                      >
                        <Box
                          sx={{
                            width: 6,
                            height: 6,
                            borderRadius: '50%',
                            backgroundColor: typeConfig.color,
                            mt: 1,
                            flexShrink: 0
                          }}
                        />
                        <Typography variant="body1" sx={{ lineHeight: 1.6 }}>
                          {change}
                        </Typography>
                      </Box>
                    ))}
                  </Box>
                </CardContent>
              </Card>
            );
          })}
        </Box>

        {/* Footer Message */}
        <Box 
          sx={{ 
            textAlign: 'center', 
            mt: 6, 
            p: 4, 
            backgroundColor: alpha(theme.palette.primary.main, 0.05),
            borderRadius: 3,
            border: `1px solid ${alpha(theme.palette.primary.main, 0.1)}`
          }}
        >          <UpdateIcon sx={{ fontSize: 48, color: 'primary.main', mb: 2 }} />
          <Typography variant="h6" gutterBottom fontWeight="600">
            {t('pages:changelog.footerTitle')}
          </Typography>
          <Typography variant="body1" color="text.secondary">
            {t('pages:changelog.footerMessage')}
          </Typography>
        </Box>
      </Container>
    </Box>
  );
};

export default Changelog;

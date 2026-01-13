import React from 'react';
import { Box, Chip, Typography } from '@mui/material';
import { Storage, CloudOff } from '@mui/icons-material';

const ModeIndicator: React.FC = () => {
  const useMockData = import.meta.env.VITE_USE_MOCK_DATA === 'true';
  const devMode = import.meta.env.VITE_DEV_MODE === 'true';
  
  if (!devMode) {
    return null; // Don't show in production
  }

  return (
    <Box
      sx={{
        position: 'fixed',
        top: 16,
        right: 16,
        zIndex: 9999,
        display: 'flex',
        gap: 1,
        flexDirection: 'column',
        alignItems: 'flex-end'
      }}
    >
      {useMockData ? (
        <Chip
          icon={<CloudOff />}
          label="Mock Mode"
          color="warning"
          variant="filled"
          size="small"
          sx={{ 
            fontSize: '0.75rem',
            fontWeight: 'bold',
            boxShadow: 2 
          }}
        />
      ) : (
        <Chip
          icon={<Storage />}
          label="API Mode"
          color="success"
          variant="filled"
          size="small"
          sx={{ 
            fontSize: '0.75rem',
            fontWeight: 'bold',
            boxShadow: 2 
          }}
        />
      )}
      
      <Typography 
        variant="caption" 
        sx={{ 
          opacity: 0.7,
          fontSize: '0.7rem',
          textAlign: 'right'
        }}
      >
        {import.meta.env.MODE} mode
      </Typography>
    </Box>
  );
};

export default ModeIndicator;

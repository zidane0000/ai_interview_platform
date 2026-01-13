import React from 'react';
import { useTranslation } from 'react-i18next';
import {
  IconButton,
  Menu,
  MenuItem,
  ListItemIcon,
  ListItemText,
  Tooltip
} from '@mui/material';
import {
  Language as LanguageIcon,
  Check as CheckIcon
} from '@mui/icons-material';

const LanguageSwitcher: React.FC = () => {
  const { i18n, t } = useTranslation();
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);

  const languages = [
    { code: 'en', name: t('common:languages.en'), flag: '🇺🇸' },
    { code: 'zh-TW', name: t('common:languages.zh-TW'), flag: '🇹🇼' }
  ];

  const handleClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleLanguageChange = (languageCode: string) => {
    i18n.changeLanguage(languageCode);
    handleClose();
  };
  const currentLanguage = languages.find(lang => lang.code === i18n.language);

  return (
    <>
      <Tooltip title="Switch Language">
        <IconButton
          onClick={handleClick}
          size="small"
          sx={{ ml: 2 }}
          aria-controls={open ? 'language-menu' : undefined}
          aria-haspopup="true"
          aria-expanded={open ? 'true' : undefined}
        >
          {currentLanguage ? (
            <span style={{ fontSize: '18px' }}>{currentLanguage.flag}</span>
          ) : (
            <LanguageIcon />
          )}
        </IconButton>
      </Tooltip>
      <Menu
        anchorEl={anchorEl}
        id="language-menu"
        open={open}
        onClose={handleClose}
        onClick={handleClose}
        transformOrigin={{ horizontal: 'right', vertical: 'top' }}
        anchorOrigin={{ horizontal: 'right', vertical: 'bottom' }}
      >
        {languages.map((language) => (
          <MenuItem
            key={language.code}
            onClick={() => handleLanguageChange(language.code)}
            selected={language.code === i18n.language}
          >
            <ListItemIcon>
              {language.code === i18n.language ? (
                <CheckIcon fontSize="small" />
              ) : (
                <span style={{ fontSize: '16px', marginLeft: '4px' }}>
                  {language.flag}
                </span>
              )}
            </ListItemIcon>
            <ListItemText>{language.name}</ListItemText>
          </MenuItem>
        ))}
      </Menu>
    </>
  );
};

export default LanguageSwitcher;

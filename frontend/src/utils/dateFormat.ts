/**
 * Utility function for locale-aware date formatting
 */

export const formatDate = (dateString: string, locale: string = 'en'): string => {
  const localeMap: Record<string, string> = {
    'en': 'en-US',
    'zh-TW': 'zh-TW'
  };

  const targetLocale = localeMap[locale] || 'en-US';

  return new Date(dateString).toLocaleDateString(targetLocale, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
};

export const formatDateShort = (dateString: string, locale: string = 'en'): string => {
  const localeMap: Record<string, string> = {
    'en': 'en-US',
    'zh-TW': 'zh-TW'
  };

  const targetLocale = localeMap[locale] || 'en-US';

  return new Date(dateString).toLocaleDateString(targetLocale, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
};

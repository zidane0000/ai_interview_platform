import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

// 導入語言資源
import enCommon from '../locales/en/common.json';
import enPages from '../locales/en/pages.json';
import enInterview from '../locales/en/interview.json';

import zhTWCommon from '../locales/zh-TW/common.json';
import zhTWPages from '../locales/zh-TW/pages.json';
import zhTWInterview from '../locales/zh-TW/interview.json';

const resources = {
  en: {
    common: enCommon,
    pages: enPages,
    interview: enInterview,
  },
  'zh-TW': {
    common: zhTWCommon,
    pages: zhTWPages,
    interview: zhTWInterview,
  },
};

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: 'en',
    defaultNS: 'common',
    
    // 偵測選項
    detection: {
      order: ['localStorage', 'navigator', 'htmlTag'],
      lookupLocalStorage: 'i18nextLng',
      caches: ['localStorage'],
    },

    interpolation: {
      escapeValue: false, // React 已經安全
    },

    // 開發模式設定
    debug: import.meta.env.DEV,
  });

export default i18n;

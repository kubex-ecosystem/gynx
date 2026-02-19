import { enUS } from './en-US';
import { ptBR } from './pt-BR';
import { LocaleTranslations, SupportedLocale } from './types';

export const translations: Record<SupportedLocale, LocaleTranslations> = {
  'en-US': enUS,
  'pt-BR': ptBR
};

export * from './types';
export { enUS, ptBR };

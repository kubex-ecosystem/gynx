# Internationalization (i18n) Structure

This project uses TypeScript-based internationalization instead of JSON files for better type safety and developer experience.

## Structure

```text
frontend/
├── locales/
│   ├── types.ts              # TypeScript interfaces for translations
│   ├── index.ts              # Main export file
│   ├── en-US/                # English translations
│   │   ├── index.ts          # English locale export
│   │   ├── common.ts         # Common translations
│   │   ├── analysis.ts       # Analysis-related translations
│   │   ├── chat.ts           # Chat translations
│   │   ├── dashboard.ts      # Dashboard translations
│   │   ├── example.ts        # Example mode translations
│   │   ├── input.ts          # Input form translations
│   │   ├── kanban.ts         # Kanban board translations
│   │   ├── landing.ts        # Landing page translations
│   │   └── settings.ts       # Settings translations
│   └── pt-BR/                # Portuguese (Brazil) translations
│       └── ... (same structure as en-US)
```

## Usage

### Using translations in components

```typescript
import { useTranslation } from '../hooks/useTranslation';

const MyComponent = () => {
  // Single namespace
  const { t } = useTranslation('common');

  // Multiple namespaces
  const { t } = useTranslation(['common', 'analysis']);

  // Usage with type checking
  const title = t('header.title');
  const message = t('analysis:results.title', { projectName: 'My Project' });

  return <div>{title}</div>;
};
```

### Adding new translations

1. **Update types**: Add new translation keys to the appropriate interface in `types.ts`
2. **Add translations**: Implement the translations in both `en-US` and `pt-BR` folders
3. **Type safety**: TypeScript will enforce that all required keys are present

### Benefits of TypeScript-based i18n

1. **Type Safety**: Compile-time checking of translation keys
2. **Better DX**: IDE autocomplete and error detection
3. **Performance**: No network requests for translations
4. **Maintainability**: Easier refactoring and missing key detection
5. **Bundle Optimization**: Only used translations are included in builds

## Migration from JSON

The previous JSON-based system in `frontend/public/locales/` has been replaced with TypeScript modules. This provides:

- Immediate loading (no async fetch required)
- Type checking for translation keys
- Better tree shaking and bundle optimization
- Easier maintenance and refactoring

## Adding New Languages

To add a new language:

1. Create a new folder in `locales/` (e.g., `es-ES/`)
2. Implement all translation modules following the type interfaces
3. Export the locale in the main `index.ts`
4. Update the `SupportedLocale` type in `types.ts`
5. Update the `LanguageContext` to support the new locale

## Performance Considerations

- Translations are statically imported and bundled
- Only the required translations for the current route are loaded
- TypeScript tree shaking eliminates unused translations
- No runtime JSON parsing or network requests

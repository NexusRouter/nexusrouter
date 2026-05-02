import js from '@eslint/js'
import react from 'eslint-plugin-react'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import globals from 'globals'
import tseslint from 'typescript-eslint'

const jsxRuntimeRules = react.configs.flat['jsx-runtime'].rules
const jsxRuntimeParserOptions =
  react.configs.flat['jsx-runtime'].languageOptions.parserOptions

export default tseslint.config(
  {
    ignores: [
      'dist/**',
      'coverage/**',
      '.vite/**',
      '**/.vite/**',
      'node_modules/**',
    ],
  },
  {
    files: ['src/**/*.{ts,tsx}', '*.{ts,tsx}', 'vite.config.ts', 'vitest.config.ts', 'vitest.setup.ts'],
    extends: [js.configs.recommended, ...tseslint.configs.recommended],
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
      parserOptions: {
        ecmaFeatures: { jsx: true },
        ...jsxRuntimeParserOptions,
      },
    },
    plugins: {
      react,
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
    },
    settings: {
      react: {
        version: 'detect',
      },
    },
    rules: {
      // --- 与 typescript-eslint / TS 分工：避免和 @typescript-eslint/* 重复告警 ---
      'no-unused-vars': 'off',
      'react/prop-types': 'off',
      'react/require-default-props': 'off',
      'react/jsx-filename-extension': [
        'warn',
        { extensions: ['.tsx', '.jsx'] },
      ],
      ...jsxRuntimeRules,
      ...reactHooks.configs.recommended.rules,
      'react-refresh/only-export-components': [
        'warn',
        { allowConstantExport: true },
      ],
    },
  },
)

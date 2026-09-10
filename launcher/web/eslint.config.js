// @ts-check
const eslint = require('@eslint/js');
const { defineConfig } = require('eslint/config');
const tseslint = require('typescript-eslint');
const angular = require('angular-eslint');

module.exports = defineConfig([
  {
    files: ['**/*.ts'],
    ignores: ['src/gen/**'],
    extends: [
      eslint.configs.recommended,
      ...tseslint.configs.recommendedTypeChecked,
      ...tseslint.configs.stylistic,
      ...angular.configs.tsRecommended,
    ],
    languageOptions: {
      parserOptions: {
        // Type-aware linting: powers the no-unsafe-* family so `any` cannot
        // leak in from untyped values, not just be written explicitly.
        projectService: true,
      },
    },
    processor: angular.processInlineTemplates,
    rules: {
      '@angular-eslint/directive-selector': [
        'error',
        { type: 'attribute', prefix: 'app', style: 'camelCase' },
      ],
      '@angular-eslint/component-selector': [
        'error',
        { type: 'element', prefix: 'app', style: 'kebab-case' },
      ],
      // Keep components and pages small: a large file means more code in a
      // single lazy chunk. Split big views into smaller components — they
      // stay lazy-loaded, so the browser downloads and parses less up front.
      'max-lines': ['error', { max: 200, skipBlankLines: true, skipComments: true }],
      '@typescript-eslint/no-explicit-any': 'error',
      // Signals all the way: no async/await and no effect() in app code. Data
      // loads through rxResource/toSignal, side effects run through RxJS
      // operators on an explicit stream, and state is derived with
      // computed/linkedSignal. The one sanctioned Promise boundary is the
      // ConnectRPC transport interceptor, whose Interceptor type is
      // Promise-based; it is carved out below.
      'no-restricted-syntax': [
        'error',
        {
          selector: 'AwaitExpression',
          message:
            'No async/await in Angular code. Return/consume Observables (rxResource, toSignal, observable-returning services). Sanctioned boundary: the ConnectRPC transport.',
        },
        {
          selector: 'FunctionExpression[async=true]',
          message:
            'No async methods/functions. Return an Observable and let signals/rxResource drive. Sanctioned boundary: the ConnectRPC transport.',
        },
        {
          selector: 'FunctionDeclaration[async=true]',
          message: 'No async functions. Return an Observable instead.',
        },
        {
          selector: 'ArrowFunctionExpression[async=true]',
          message: 'No async arrow functions. Return an Observable instead.',
        },
        {
          selector: "CallExpression[callee.name='effect']",
          message:
            'No effect(). Derive with computed/linkedSignal, load with rxResource, run side effects through RxJS operators on an explicit stream.',
        },
        {
          selector: 'TSUnknownKeyword',
          message:
            'No bare `unknown` — use a concrete type. The only sanctioned unknowns (caught error values, the rxResource error signal) opt out per line with `// eslint-disable-next-line no-restricted-syntax -- <reason>`.',
        },
      ],
      // No parent-relative imports (`../`). They obscure where a symbol lives
      // and break on every move. Use the path aliases instead; same-directory
      // `./` imports stay allowed. Plus: never bridge an observable to a
      // Promise, and never reach for Reactive forms — signal forms only.
      '@typescript-eslint/no-restricted-imports': [
        'error',
        {
          paths: [
            {
              name: 'rxjs',
              importNames: ['firstValueFrom', 'lastValueFrom'],
              message:
                'Do not convert observables to Promises. Consume the observable via rxResource/toSignal, or return it.',
            },
            {
              name: '@angular/forms',
              importNames: [
                'ReactiveFormsModule',
                'FormGroup',
                'FormControl',
                'FormArray',
                'FormBuilder',
              ],
              message:
                'Use signal forms (@angular/forms/signals: form() + [field]) instead of Reactive forms.',
            },
          ],
          patterns: [
            {
              group: ['../**'],
              message: 'Use path aliases (@app/*, @gen/*) instead of parent-relative imports.',
            },
          ],
        },
      ],
    },
  },
  {
    // The one sanctioned Promise boundary: ConnectRPC's Interceptor type is
    // Promise-based, so async is correct here and only here. effect() stays
    // banned everywhere.
    files: ['src/app/transport/connect-transport.ts'],
    rules: {
      'no-restricted-syntax': [
        'error',
        {
          selector: "CallExpression[callee.name='effect']",
          message:
            'No effect(). Derive with computed/linkedSignal, load with rxResource, run side effects through RxJS operators on an explicit stream.',
        },
        {
          selector: 'TSUnknownKeyword',
          message:
            'No bare `unknown` — use a concrete type (per-line opt-out with `// eslint-disable-next-line no-restricted-syntax -- <reason>`).',
        },
      ],
    },
  },
  {
    // A spec that fakes a transport has to describe what the real one is handed:
    // a Connect Transport takes the service descriptor, the method and the
    // message as separate arguments, and typing them properly here would be
    // rebuilding Connect's own generics to say "I ignored these". async is the
    // shape a fake Promise-based API is written in, for the same reason the
    // transport itself is carved out above.
    files: ['**/*.spec.ts'],
    rules: {
      'no-restricted-syntax': [
        'error',
        {
          selector: "CallExpression[callee.name='effect']",
          message:
            'No effect(). Derive with computed/linkedSignal, load with rxResource, run side effects through RxJS operators on an explicit stream.',
        },
      ],
    },
  },
  {
    // The browser tests are not Angular code. Playwright's whole API is
    // Promise-based and reads as async/await; there is no change detection here
    // to keep off the microtask queue and nothing to hold in a signal. The rest
    // of the rules stay.
    files: ['e2e/**/*.ts', 'playwright.config.ts'],
    rules: {
      'no-restricted-syntax': [
        'error',
        {
          selector: 'TSUnknownKeyword',
          message: 'No bare `unknown` — use a concrete type.',
        },
      ],
    },
  },
  {
    files: ['**/*.html'],
    extends: [...angular.configs.templateRecommended, ...angular.configs.templateAccessibility],
    rules: {},
  },
]);

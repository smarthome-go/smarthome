module.exports = {
    'env': {
        'browser': true,
        'es2021': true,
    },
    overrides: [
        {
            files: ['*.svelte'],
            processor: 'svelte3/svelte3'
        }
    ],
    'parser': '@typescript-eslint/parser',
    'parserOptions': {
        'ecmaVersion': 'latest',
        tsconfigRootDir: __dirname,
        project: ['./tsconfig.json'],
        extraFileExtensions: ['.svelte'],
    },
    'plugins': [
        '@typescript-eslint',
        'svelte3',
    ],
    'rules': {
        '@typescript-eslint/ban-ts-comment': 'off',
    },
    settings: {
        'svelte3/typescript': () => require('typescript'),
    },
    extends: [
        'eslint:recommended',
        'plugin:@typescript-eslint/recommended',
    ]
};

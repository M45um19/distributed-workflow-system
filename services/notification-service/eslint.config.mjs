import pluginJs from "@eslint/js";
import eslintConfigPrettier from "eslint-config-prettier";
import importPlugin from "eslint-plugin-import";
import globals from "globals";
import tseslint from "typescript-eslint";

/** @type {import('eslint').Linter.Config[]} */
export default [
  { 
    files: ["**/*.{js,mjs,cjs,ts}"],
    plugins: {
      import: importPlugin,
    },
  },
  { languageOptions: { globals: globals.node } },
  
  pluginJs.configs.recommended,
  
  ...tseslint.configs.strict,
  ...tseslint.configs.stylistic,

  {
    rules: {
      "no-console": "warn",
      "no-unused-vars": "off",
      "@typescript-eslint/no-unused-vars": ["error", { "argsIgnorePattern": "^_" }],
      "@typescript-eslint/no-explicit-any": "error",
      "@typescript-eslint/consistent-type-definitions": ["error", "interface"],
      
      "import/order": [
        "error",
        {
          "groups": ["builtin", "external", "internal", "parent", "sibling", "index"],
          "newlines-between": "always",
          "alphabetize": { "order": "asc", "caseInsensitive": true }
        }
      ],
      
      "eqeqeq": ["error", "always"],
      "no-return-await": "error",
    }
  },

  eslintConfigPrettier,

  {
    ignores: ["dist/", "node_modules/", "proto/", "coverage/"]
  }
];
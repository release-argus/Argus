module.exports = {
  "root": true,
  "parser": "@typescript-eslint/parser",
  "extends": [
    'plugin:react/recommended',
    'plugin:react/jsx-runtime',
    "plugin:@typescript-eslint/recommended"
  ],
  "rules": {
    "@typescript-eslint/no-unused-vars": "error",
    "no-unused-vars": "off"
  }
}

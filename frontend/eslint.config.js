import { dirname } from "path";
import { fileURLToPath } from "url";
import { FlatCompat } from "@eslint/eslintrc";
import typescriptEslint from "@typescript-eslint/eslint-plugin";
import tsParser from "@typescript-eslint/parser";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const compat = new FlatCompat({
  baseDirectory: __dirname,
});

const eslintConfig = [
  ...compat.extends("next/core-web-vitals"),
  {
    plugins: {
      "@typescript-eslint": typescriptEslint,
    },
    languageOptions: {
      parser: tsParser,
    },
    rules: {
      // 禁止裸用 fetch 调业务接口，必须走 apiClient（统一鉴权 + 多租户头）。
      // 两条规则分别覆盖字符串字面量与模板字符串，防止插值绕过（曾因此漏检）。
      "no-restricted-syntax": ["error",
        {
          selector: "CallExpression[callee.name='fetch'][arguments.0.value=/^\\/?api/]",
          message: "业务请求请走 apiClient，勿裸用 fetch"
        },
        {
          selector: "CallExpression[callee.name='fetch'] > TemplateLiteral:first-child",
          message: "业务请求请走 apiClient，勿裸用 fetch（含模板字符串拼接）"
        }
      ],
    },
  },
];

export default eslintConfig;

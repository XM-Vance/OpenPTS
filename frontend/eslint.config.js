// ESLint 9 flat config。
// Next 16 起 eslint-config-next 提供原生 flat config，直接 import 展开即可
// （旧版经 @eslint/eslintrc FlatCompat 加载 next/core-web-vitals，Next 16 下会触发
//  "Converting circular structure to JSON" 报错）。
import nextCoreWebVitals from "eslint-config-next/core-web-vitals";

const eslintConfig = [
  {
    // 迁移到 ESLint CLI 后，next lint 内置的忽略规则不再生效，这里显式声明：
    // 跳过构建产物、依赖与配置文件（与原 `next lint` 行为一致）。
    ignores: [
      ".next/**",
      "node_modules/**",
      "out/**",
      "build/**",
      "dist/**",
      "postcss.config.mjs",
      "next.config.*",
      "eslint.config.js",
    ],
  },
  // next/core-web-vitals：原生 flat config，已为 .ts/.tsx 配好 @typescript-eslint 插件与解析器。
  ...nextCoreWebVitals,
  {
    rules: {
      // eslint-config-next 16 默认启用了 React Compiler 系 react-hooks 规则（plugin v6，
      // purity / set-state-in-effect / immutability 等），会把大量存量组件模式判为 error。
      // 属独立重构范畴、有行为风险，本次仅做 Next 16 升级，先关闭以保持与升级前等价的检查
      // 范围，后续单独立项整改（见 TODO：react-hooks v6 规则整改）。
      "react-hooks/purity": "off",
      "react-hooks/set-state-in-effect": "off",
      "react-hooks/set-state-in-render": "off",
      "react-hooks/preserve-manual-memoization": "off",
      "react-hooks/immutability": "off",
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

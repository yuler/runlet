# Runlet Design

Runlet 的视觉方向应保持简单、清晰、偏开发者工具感：灰白、克制、可快速扫描。

参考：[JimLiu/baoyu-design](https://github.com/JimLiu/baoyu-design) 的思路，先给足设计上下文，再产出可落地的界面。

## Colors

| 用途 | 颜色 | 说明 |
| --- | --- | --- |
| Background | `#FFFFFF` | 主背景，保持干净留白 |
| Subtle Background | `#FAFAFA` | 页面底色、弱分区 |
| Surface | `#FFFFFF` | 卡片、面板、输入框背景 |
| Surface Muted | `#F5F5F5` | 悬浮、禁用、浅色块 |
| Text | `#171717` | 主标题、正文、关键 UI |
| Muted Text | `#525252` | 次级文字 |
| Subtle Text | `#737373` | 说明文字、辅助信息 |
| Accent | `#171717` | 主操作、当前状态、关键强调 |
| Accent Soft | `#EDEDED` | 轻强调背景 |
| Success | `#86EFAC` | 成功、运行中、通过 |
| Warning | `#FDE68A` | 等待、提醒 |
| Error | `#FCA5A5` | 错误、失败 |
| Border | `#E5E5E5` | 默认边框 |
| Divider | `#F0F0F0` | 轻分割线 |

## Typography

- Font: `Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif`
- Title: 32px / 1.1 / weight 700+
- Body: 16px / 1.6 / weight 400
- Caption: 14px / 1.4 / weight 400
- Letter spacing: `0`

## Shape

- Card radius: `8px`
- Logo tile radius: `36px`
- Button radius: `8px`
- Border: 1px solid `Border`

## Spacing

- Page padding: `40px` desktop, `20px` mobile
- Section gap: `28px`
- Grid gap: `18px`
- Compact padding: `12px 14px`

## Usage

- Prefer white and near-white backgrounds with dark text.
- Use black as the main accent; avoid bright decorative colors.
- Keep layouts calm and functional; avoid decorative gradients and overly rounded cards.
- When unsure, follow Vercel / Notion style: restrained borders, quiet surfaces, strong typography.
